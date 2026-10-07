package takigo

import (
	"encoding/binary"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// Drop is what another application dropped on a window registered with
// App.OnDrop.
type Drop struct {
	// Window is the registered window the drop landed on, and X, Y the
	// pointer position inside it.
	Window *window.Window
	X, Y   int
	// Files holds the paths of dropped files, and Text the dropped text
	// when there were none.
	Files []string
	Text  string
}

// xdndVersion is the version of the XDND protocol this side speaks.
const xdndVersion = 5

// dnd is the drop-target side of XDND, the drag-and-drop protocol of X11
// desktops. Tk has none (tkdnd is an extension). Nothing arrives on Windows
// and macOS, whose backends do not deliver these messages.
type dnd struct {
	app      *App
	handlers map[*window.Window]func(Drop)

	aware, typeList, selection                platform.AtomID
	enter, position, status, leave, drop, fin platform.AtomID
	actionCopy                                platform.AtomID
	accepted                                  []platform.AtomID // data types, best first

	// The drag in progress over one of our windows.
	source  platform.WindowID
	offered []platform.AtomID
	target  *window.Window
	x, y    int

	// The drag this App started, if any.
	out *outgoingDrag
}

// DragData is what App.StartDrag offers: files, or text when there are none.
type DragData struct {
	Files []string
	Text  string
}

// outgoingDrag is the source side of one drag.
type outgoingDrag struct {
	from     platform.WindowID // our toplevel, the XDND source window
	types    []platform.AtomID
	over     platform.WindowID // the XdndAware window under the pointer
	accepted bool              // its last XdndStatus accepted the drop
	bindings []event.BindingID
	done     func(dropped bool)
}

// OnDrop makes w accept text and files dragged onto it from other
// applications and calls handler with each drop; a nil handler stops that.
// With several registered windows under the pointer, the innermost gets
// the drop. It works on X11 desktops (XDND).
func (a *App) OnDrop(w window.Windower, handler func(Drop)) {
	if a.dnd == nil {
		a.dnd = newDnd(a)
	}
	win := w.Window()
	if handler == nil {
		delete(a.dnd.handlers, win)
		return
	}
	if _, ok := a.dnd.handlers[win]; !ok {
		win.OnDestroy(func() { delete(a.dnd.handlers, win) })
	}
	a.dnd.handlers[win] = handler
	// The source looks for XdndAware on the toplevel under the pointer.
	if top := window.Toplevel(win); top != nil && top.PlatformID != 0 {
		version := binary.NativeEndian.AppendUint32(nil, xdndVersion)
		a.display.Server.ChangeProperty(top.PlatformID, a.dnd.aware, a.display.Server.Atoms().Atom,
			32, platform.PropModeReplace, version, 1)
	}
}

func newDnd(a *App) *dnd {
	s := a.display.Server
	atom := func(name string) platform.AtomID { return s.InternAtom(name, false) }
	return &dnd{
		app:        a,
		handlers:   make(map[*window.Window]func(Drop)),
		aware:      atom("XdndAware"),
		typeList:   atom("XdndTypeList"),
		selection:  atom("XdndSelection"),
		enter:      atom("XdndEnter"),
		position:   atom("XdndPosition"),
		status:     atom("XdndStatus"),
		leave:      atom("XdndLeave"),
		drop:       atom("XdndDrop"),
		fin:        atom("XdndFinished"),
		actionCopy: atom("XdndActionCopy"),
		accepted: []platform.AtomID{
			atom("text/uri-list"), atom("UTF8_STRING"), atom("text/plain;charset=utf-8"),
			atom("text/plain"), s.Atoms().String,
		},
	}
}

// handle processes an XDND client message sent to one of our toplevels and
// reports whether ev was one.
func (d *dnd) handle(ev *event.Event) bool {
	s := d.app.display.Server
	data := ev.MessageData
	switch ev.MessageType {
	case d.enter:
		d.source = platform.WindowID(data[0])
		d.target = nil
		d.offered = d.offered[:0]
		if data[1]&1 != 0 {
			// More than three types: they are in a property on the source.
			raw, _, _ := s.GetWindowProperty(d.source, d.typeList, 0, 1024, false)
			for i := 0; i+4 <= len(raw); i += 4 {
				d.offered = append(d.offered, platform.AtomID(binary.NativeEndian.Uint32(raw[i:])))
			}
		} else {
			for _, t := range data[2:5] {
				if t != 0 {
					d.offered = append(d.offered, platform.AtomID(t))
				}
			}
		}
	case d.position:
		if platform.WindowID(data[0]) != d.source {
			return true
		}
		rootX, rootY := int(data[2]>>16&0xffff), int(data[2]&0xffff)
		d.target, d.x, d.y = d.windowAt(ev.Window, rootX, rootY)
		accept, action := int64(0), int64(0)
		if d.target != nil && d.bestType() != 0 {
			accept, action = 1, int64(d.actionCopy)
		}
		// Bit 1 asks for a position message on every move, since which of
		// our windows is under the pointer may change.
		s.SendClientMessage(d.source, d.source, d.status, int64(ev.Window), accept|2, 0, 0, action)
	case d.status:
		if d.out != nil && platform.WindowID(data[0]) == d.out.over {
			d.out.accepted = data[1]&1 != 0
		}
	case d.fin:
		if d.out != nil && d.out.over == 0 {
			d.endDrag(data[1]&1 != 0)
		}
	case d.leave:
		d.source, d.target = 0, nil
	case d.drop:
		d.finishDrop(ev.Window, platform.Timestamp(data[2]))
	default:
		return false
	}
	return true
}

func (d *dnd) bestType() platform.AtomID {
	for _, want := range d.accepted {
		if slices.Contains(d.offered, want) {
			return want
		}
	}
	return 0
}

// windowAt returns the innermost registered window of the toplevel with
// platform window top that contains the root position, and the position
// inside it.
func (d *dnd) windowAt(top platform.WindowID, rootX, rootY int) (*window.Window, int, int) {
	disp := d.app.display
	var best *window.Window
	var bx, by int
	for w := range d.handlers {
		if w.PlatformID == 0 || w.IsDestroyed() {
			continue
		}
		if t := window.Toplevel(w); t == nil || t.PlatformID != top {
			continue
		}
		x, y := disp.Server.TranslateCoordinates(disp.RootWindow, w.PlatformID, rootX, rootY)
		if x < 0 || y < 0 || x >= w.Width || y >= w.Height {
			continue
		}
		if best == nil || w.Width*w.Height < best.Width*best.Height {
			best, bx, by = w, x, y
		}
	}
	return best, bx, by
}

func (d *dnd) finishDrop(top platform.WindowID, ts platform.Timestamp) {
	s := d.app.display.Server
	source, target, x, y := d.source, d.target, d.x, d.y
	typ := d.bestType()
	d.source, d.target = 0, nil
	done := func(accepted bool) {
		ok, action := int64(0), int64(0)
		if accepted {
			ok, action = 1, int64(d.actionCopy)
		}
		s.SendClientMessage(source, source, d.fin, int64(top), ok, action, 0, 0)
		s.Flush()
	}
	handler := d.handlers[target]
	if target == nil || handler == nil || typ == 0 {
		done(false)
		return
	}
	d.app.selMgr.RequestTarget(d.selection, typ, top, ts, func(data []byte) {
		if data == nil {
			done(false)
			return
		}
		drop := Drop{Window: target, X: x, Y: y}
		if typ == d.accepted[0] {
			drop.Files = uriListFiles(string(data))
		}
		if len(drop.Files) == 0 {
			drop.Text = string(data)
		}
		done(true)
		handler(drop)
	})
}

// uriListFiles returns the local paths in a text/uri-list (RFC 2483):
// one URI per CRLF-terminated line, "#" lines being comments.
func uriListFiles(list string) []string {
	var files []string
	for line := range strings.SplitSeq(list, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		u, err := url.Parse(line)
		if err != nil || u.Scheme != "file" || u.Path == "" {
			continue
		}
		files = append(files, u.Path)
	}
	return files
}

// StartDrag starts dragging data out of w to wherever the user releases the
// mouse button: another application or one of this App's own OnDrop
// windows. Call it from a handler while a button is held on w, typically
// on the first Motion after a ButtonPress. done, which may be nil, is
// called when the drag ends, with whether a target took the data. It works
// on X11 desktops (XDND); elsewhere done is called with false at once.
func (a *App) StartDrag(w window.Windower, data DragData, done func(dropped bool)) {
	if a.dnd == nil {
		a.dnd = newDnd(a)
	}
	a.dnd.startDrag(w.Window(), data, done)
}

func (d *dnd) startDrag(w *window.Window, data DragData, done func(bool)) {
	if done == nil {
		done = func(bool) {}
	}
	top := window.Toplevel(w)
	if d.out != nil || top == nil || top.PlatformID == 0 || (len(data.Files) == 0 && data.Text == "") {
		done(false)
		return
	}
	s := d.app.display.Server
	formats := map[platform.AtomID][]byte{}
	text := data.Text
	if len(data.Files) > 0 {
		var list strings.Builder
		for _, f := range data.Files {
			list.WriteString((&url.URL{Scheme: "file", Path: f}).String() + "\r\n")
		}
		formats[d.accepted[0]] = []byte(list.String())
		text = strings.Join(data.Files, "\n")
	}
	for _, t := range d.accepted[1:4] {
		formats[t] = []byte(text)
	}
	out := &outgoingDrag{from: top.PlatformID, done: done}
	for _, t := range d.accepted {
		if _, ok := formats[t]; ok {
			out.types = append(out.types, t)
		}
	}
	d.app.selMgr.OwnFormats(d.selection, top.PlatformID, formats, platform.CurrentTime)
	// More than three types go in a property the target reads.
	list := make([]byte, 0, 4*len(out.types))
	for _, t := range out.types {
		list = binary.NativeEndian.AppendUint32(list, uint32(t))
	}
	s.ChangeProperty(top.PlatformID, d.typeList, s.Atoms().Atom, 32, platform.PropModeReplace, list, len(out.types))
	d.out = out

	// While the button is held the X server sends every pointer event to
	// the window it was pressed in, with root coordinates.
	disp := d.app.dispatcher
	out.bindings = append(out.bindings,
		disp.BindGlobal(event.MotionMask, func(ev *event.Event) {
			if ev.Type == event.MotionType && d.out == out {
				d.dragMove(ev.RootX, ev.RootY, ev.Time)
			}
		}),
		disp.BindGlobal(event.ButtonReleaseMask, func(ev *event.Event) {
			if ev.Type == event.ButtonReleaseType && d.out == out {
				d.dragMove(ev.RootX, ev.RootY, ev.Time)
				d.dragRelease(ev.Time)
			}
		}))
}

// awareWindowAt returns the innermost window under the root position that
// has the XdndAware property, and the XDND version it speaks.
func (d *dnd) awareWindowAt(x, y int) (platform.WindowID, int64) {
	s := d.app.display.Server
	win := d.app.display.RootWindow
	for range 32 { // deeper than any real window tree
		child := s.ChildAt(win, x, y)
		if child == 0 {
			break
		}
		win = child
		if raw, _, _ := s.GetWindowProperty(win, d.aware, 0, 1, false); len(raw) >= 4 {
			return win, min(int64(binary.NativeEndian.Uint32(raw)), xdndVersion)
		}
	}
	return 0, 0
}

func (d *dnd) dragMove(x, y int, ts platform.Timestamp) {
	s := d.app.display.Server
	out := d.out
	over, version := d.awareWindowAt(x, y)
	if over != out.over {
		if out.over != 0 {
			s.SendClientMessage(out.over, out.over, d.leave, int64(out.from), 0, 0, 0, 0)
		}
		out.over, out.accepted = over, false
		if over != 0 {
			var first [3]int64
			for i := range min(len(out.types), 3) {
				first[i] = int64(out.types[i])
			}
			more := int64(0)
			if len(out.types) > 3 {
				more = 1
			}
			s.SendClientMessage(over, over, d.enter, int64(out.from), version<<24|more, first[0], first[1], first[2])
		}
	}
	if over != 0 {
		s.SendClientMessage(over, over, d.position, int64(out.from), 0,
			int64(x&0xffff)<<16|int64(y&0xffff), int64(ts), int64(d.actionCopy))
	}
	s.Flush()
}

func (d *dnd) dragRelease(ts platform.Timestamp) {
	s := d.app.display.Server
	out := d.out
	for _, id := range out.bindings {
		d.app.dispatcher.UnbindID(id)
	}
	out.bindings = nil
	over := out.over
	// A target gets the drop on its word that it will take it. Its status
	// for the last position may still be on its way, so the release waits
	// for it, with a limit.
	finish := func() {
		if d.out != out {
			return
		}
		if over == 0 || !out.accepted {
			if over != 0 {
				s.SendClientMessage(over, over, d.leave, int64(out.from), 0, 0, 0, 0)
				s.Flush()
			}
			d.endDrag(false)
			return
		}
		out.over = 0 // waiting for XdndFinished
		s.SendClientMessage(over, over, d.drop, int64(out.from), 0, int64(ts), 0, 0)
		s.Flush()
		d.app.loop.After(dragTimeout, func() {
			if d.out == out {
				d.endDrag(false)
			}
		})
	}
	if over == 0 || out.accepted {
		finish()
		return
	}
	d.app.loop.After(statusWait, finish)
}

// statusWait is how long a release waits for the target's answer to the
// last position; dragTimeout how long for it to finish the transfer.
const (
	statusWait  = 150 * time.Millisecond
	dragTimeout = 5 * time.Second
)

func (d *dnd) endDrag(dropped bool) {
	out := d.out
	if out == nil {
		return
	}
	d.out = nil
	for _, id := range out.bindings {
		d.app.dispatcher.UnbindID(id)
	}
	out.done(dropped)
}

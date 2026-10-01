package takigo

import (
	"encoding/binary"
	"net/url"
	"slices"
	"strings"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
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

	// The drag in progress.
	source  platform.WindowID
	offered []platform.AtomID
	target  *window.Window
	x, y    int
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

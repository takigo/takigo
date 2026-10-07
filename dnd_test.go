package takigo

import (
	"slices"
	"testing"
	"time"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/internal/displaylock"
	"github.com/takigo/takigo/platform"
)

func TestURIListFiles(t *testing.T) {
	list := "# a comment\r\nfile:///tmp/a%20b.txt\r\nfile://localhost/etc/hosts\r\nhttps://example.com/x\r\n\r\nnot a uri at all\r\n"
	want := []string{"/tmp/a b.txt", "/etc/hosts"}
	if got := uriListFiles(list); !slices.Equal(got, want) {
		t.Errorf("uriListFiles = %q, want %q", got, want)
	}
	if got := uriListFiles(""); len(got) != 0 {
		t.Errorf("empty list gave %q", got)
	}
}

// settle waits for the display server and window manager to finish with
// Apps destroyed just before (see testutil.Settle, which this package
// cannot import).
func settle() { time.Sleep(200 * time.Millisecond) }

// rootOrigin returns where app's root window sits on the screen once the
// window manager has stopped moving it: a window manager, unlike the
// virtual server of xvfb-run, places a window after mapping it.
func rootOrigin(app *App) (x, y int) {
	d := app.display
	read := func() (int, int) {
		return d.Server.TranslateCoordinates(app.root.PlatformID, d.RootWindow, 0, 0)
	}
	x, y = read()
	for range 40 {
		time.Sleep(50 * time.Millisecond)
		nx, ny := read()
		if nx == x && ny == y {
			time.Sleep(50 * time.Millisecond)
			if nx, ny = read(); nx == x && ny == y {
				return x, y
			}
		}
		x, y = nx, ny
	}
	return x, y
}

// raise puts app's window on top, as far as the window manager allows: a
// desktop may map a new window behind the one the user is working in.
func raise(app *App) {
	app.display.Server.RaiseWindow(app.root.PlatformID)
	app.display.Server.Flush()
}

// uncoveredPoint returns a point inside app's root window, in its own
// coordinates, where it is the window a drag would find under the pointer:
// a window manager stacks the test's windows as it likes, and one App's
// window may cover part of another's.
func uncoveredPoint(t *testing.T, probe, app *App) (x, y int) {
	t.Helper()
	raise(app)
	time.Sleep(100 * time.Millisecond)
	ox, oy := rootOrigin(app)
	for y := 10; y < app.root.Height-10; y += 10 {
		for x := 10; x < app.root.Width-10; x += 10 {
			if w, _ := probe.dnd.awareWindowAt(ox+x, oy+y); w == app.root.PlatformID {
				return x, y
			}
		}
	}
	t.Skip("the window manager left no part of the window uncovered")
	return 0, 0
}

// One App plays the drag source of the XDND protocol by hand and drops a
// file list and then text on another App, a separate X client.
func TestDropFromAnotherClient(t *testing.T) {
	if !haveDisplayBackend {
		t.Skip("no display backend in this build (cgo disabled)")
	}
	displaylock.Require(t)
	settle()

	target, err := NewApp(Title("target"), Size(200, 150), Geometry("+0+0"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := NewApp(Title("source"), Size(50, 50), Geometry("+400+400"))
	if err != nil {
		t.Fatal(err)
	}

	drops := make(chan Drop, 2)
	target.OnDrop(target, func(d Drop) { drops <- d })

	srv := source.display.Server
	atom := func(name string) platform.AtomID { return srv.InternAtom(name, false) }
	type reply struct {
		typ    platform.AtomID
		accept int64
	}
	replies := make(chan reply, 8)
	source.dispatcher.BindGlobal(event.ClientMessageMask, func(ev *event.Event) {
		if ev.Type == event.ClientMessageType && (ev.MessageType == atom("XdndStatus") || ev.MessageType == atom("XdndFinished")) {
			replies <- reply{ev.MessageType, ev.MessageData[1] & 1}
		}
	})

	targetDone, sourceDone := make(chan struct{}), make(chan struct{})
	go func() { target.Run(); close(targetDone) }()
	go func() { source.Run(); close(sourceDone) }()
	defer func() {
		target.Quit()
		source.Quit()
		<-targetDone
		<-sourceDone
	}()

	var ox, oy int // where the target sits on the screen, set once it is mapped
	src, dst := int64(source.root.PlatformID), target.root.PlatformID
	drag := func(typ string, data string, x, y int) {
		source.RunOnMain(func() {
			source.selMgr.OwnFormats(atom("XdndSelection"), source.root.PlatformID,
				map[platform.AtomID][]byte{atom(typ): []byte(data)}, platform.CurrentTime)
			srv.SendClientMessage(dst, dst, atom("XdndEnter"), src, xdndVersion<<24, int64(atom(typ)), 0, 0)
			srv.SendClientMessage(dst, dst, atom("XdndPosition"), src, 0, int64((ox+x)<<16|(oy+y)), 0, int64(atom("XdndActionCopy")))
			srv.SendClientMessage(dst, dst, atom("XdndDrop"), src, 0, 0, 0, 0)
			srv.Flush()
		})
	}
	wait := func(what string) reply {
		t.Helper()
		select {
		case r := <-replies:
			return r
		case <-time.After(10 * time.Second):
			t.Fatalf("no %s from the target", what)
			return reply{}
		}
	}
	waitDrop := func() Drop {
		t.Helper()
		select {
		case d := <-drops:
			return d
		case <-time.After(10 * time.Second):
			t.Fatal("the drop handler was not called")
			return Drop{}
		}
	}

	time.Sleep(300 * time.Millisecond) // both windows mapped
	ox, oy = rootOrigin(target)

	drag("text/uri-list", "file:///tmp/a%20b.txt\r\nfile:///etc/hosts\r\n", 30, 40)
	if r := wait("XdndStatus"); r.typ != atom("XdndStatus") || r.accept != 1 {
		t.Errorf("first reply = %+v, want an accepting XdndStatus", r)
	}
	d := waitDrop()
	if !slices.Equal(d.Files, []string{"/tmp/a b.txt", "/etc/hosts"}) || d.Text != "" {
		t.Errorf("drop = %+v, want the two files", d)
	}
	if d.Window != target.root || d.X != 30 || d.Y != 40 {
		t.Errorf("drop at %d,%d on %v, want 30,40 on the target root", d.X, d.Y, d.Window.PathName)
	}
	if r := wait("XdndFinished"); r.typ != atom("XdndFinished") || r.accept != 1 {
		t.Errorf("second reply = %+v, want an accepting XdndFinished", r)
	}

	drag("UTF8_STRING", "dragged text", 5, 5)
	wait("XdndStatus")
	if d := waitDrop(); d.Text != "dragged text" || len(d.Files) != 0 {
		t.Errorf("text drop = %+v", d)
	}
	wait("XdndFinished")

	// A type the target does not take is refused, and nothing is delivered.
	drag("application/x-unknown", "zzz", 5, 5)
	if r := wait("XdndStatus"); r.accept != 0 {
		t.Errorf("an unknown type was accepted: %+v", r)
	}
	if r := wait("XdndFinished"); r.accept != 0 {
		t.Errorf("an unknown type finished as accepted: %+v", r)
	}
	select {
	case d := <-drops:
		t.Errorf("an unknown type was delivered: %+v", d)
	default:
	}
}

// A real drag: StartDrag in one App, the pointer simulated by motion and
// release events carrying root coordinates, dropped on another App.
func TestStartDragOntoAnotherClient(t *testing.T) {
	if !haveDisplayBackend {
		t.Skip("no display backend in this build (cgo disabled)")
	}
	displaylock.Require(t)
	settle()

	target, err := NewApp(Title("target"), Size(200, 150), Geometry("+0+0"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := NewApp(Title("source"), Size(60, 60), Geometry("+400+400"))
	if err != nil {
		t.Fatal(err)
	}
	drops := make(chan Drop, 4)
	target.OnDrop(target, func(d Drop) { drops <- d })
	// The source takes drops too, for a drag that ends on itself.
	own := make(chan Drop, 1)
	source.OnDrop(source, func(d Drop) { own <- d })

	targetDone, sourceDone := make(chan struct{}), make(chan struct{})
	go func() { target.Run(); close(targetDone) }()
	go func() { source.Run(); close(sourceDone) }()
	defer func() {
		target.Quit()
		source.Quit()
		<-targetDone
		<-sourceDone
	}()
	time.Sleep(300 * time.Millisecond)
	tx, ty := rootOrigin(target)
	sx, sy := rootOrigin(source)
	px, py := uncoveredPoint(t, source, target)
	qx, qy := uncoveredPoint(t, source, source)

	results := make(chan bool, 4)
	pointer := func(typ event.Type, x, y int) {
		source.dispatcher.Dispatch(&event.Event{Type: typ, Window: source.root.PlatformID, RootX: x, RootY: y})
	}
	drag := func(data DragData, path [][2]int) {
		source.RunOnMain(func() {
			source.StartDrag(source, data, func(dropped bool) { results <- dropped })
			for _, p := range path {
				pointer(event.MotionType, p[0], p[1])
			}
			last := path[len(path)-1]
			pointer(event.ButtonReleaseType, last[0], last[1])
		})
	}
	result := func() bool {
		t.Helper()
		select {
		case r := <-results:
			return r
		case <-time.After(10 * time.Second):
			t.Fatal("the drag never ended")
			return false
		}
	}

	// Across to the other client, entering it on the way.
	drag(DragData{Files: []string{"/tmp/report 1.pdf", "/etc/hosts"}}, [][2]int{{sx + qx, sy + qy}, {tx + px, ty + py}})
	if !result() {
		t.Error("a drop on an accepting window was reported as not dropped")
	}
	select {
	case d := <-drops:
		if !slices.Equal(d.Files, []string{"/tmp/report 1.pdf", "/etc/hosts"}) || d.X != px || d.Y != py {
			t.Errorf("drop = %+v, want the two files at %d,%d", d, px, py)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the target did not get the drop")
	}

	// Text.
	drag(DragData{Text: "some dragged text"}, [][2]int{{tx + px, ty + py}})
	if !result() {
		t.Error("the text drag was not dropped")
	}
	if d := <-drops; d.Text != "some dragged text" {
		t.Errorf("text drop = %+v", d)
	}

	// Released off the screen, where no window can be: nobody takes it.
	// (A point on the screen could be over a real application's window.)
	drag(DragData{Text: "nowhere"}, [][2]int{{tx + px, ty + py}, {-10, -10}})
	if result() {
		t.Error("a drag released over nothing was reported as dropped")
	}

	// Onto the App's own window.
	drag(DragData{Text: "to myself"}, [][2]int{{sx + qx, sy + qy}})
	if !result() {
		t.Error("a drag onto the App's own window was not dropped")
	}
	select {
	case d := <-own:
		if d.Text != "to myself" {
			t.Errorf("own drop = %+v", d)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the App did not get its own drop")
	}
	select {
	case d := <-drops:
		t.Errorf("the other client got a stray drop: %+v", d)
	default:
	}

	// Nothing to drag ends at once.
	source.RunOnMain(func() { source.StartDrag(source, DragData{}, func(dropped bool) { results <- dropped }) })
	if result() {
		t.Error("an empty drag was reported as dropped")
	}
}

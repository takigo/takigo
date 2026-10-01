package takigo

import (
	"os"
	"slices"
	"testing"
	"time"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/internal/displaylock"
	"github.com/msorc/takigo/platform"
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

// One App plays the drag source of the XDND protocol by hand and drops a
// file list and then text on another App, a separate X client.
func TestDropFromAnotherClient(t *testing.T) {
	if os.Getenv("DISPLAY") == "" || !haveDisplayBackend {
		t.Skip("needs an X display")
	}
	displaylock.Acquire(t)

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

	src, dst := int64(source.root.PlatformID), target.root.PlatformID
	drag := func(typ string, data string, x, y int) {
		source.RunOnMain(func() {
			source.selMgr.OwnFormats(atom("XdndSelection"), source.root.PlatformID,
				map[platform.AtomID][]byte{atom(typ): []byte(data)}, platform.CurrentTime)
			srv.SendClientMessage(dst, dst, atom("XdndEnter"), src, xdndVersion<<24, int64(atom(typ)), 0, 0)
			srv.SendClientMessage(dst, dst, atom("XdndPosition"), src, 0, int64(x<<16|y), 0, int64(atom("XdndActionCopy")))
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
	if os.Getenv("DISPLAY") == "" || !haveDisplayBackend {
		t.Skip("needs an X display")
	}
	displaylock.Acquire(t)

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
	drag(DragData{Files: []string{"/tmp/report 1.pdf", "/etc/hosts"}}, [][2]int{{410, 410}, {300, 300}, {150, 100}, {60, 70}})
	if !result() {
		t.Error("a drop on an accepting window was reported as not dropped")
	}
	select {
	case d := <-drops:
		if !slices.Equal(d.Files, []string{"/tmp/report 1.pdf", "/etc/hosts"}) || d.X != 60 || d.Y != 70 {
			t.Errorf("drop = %+v, want the two files at 60,70", d)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the target did not get the drop")
	}

	// Text.
	drag(DragData{Text: "some dragged text"}, [][2]int{{20, 20}})
	if !result() {
		t.Error("the text drag was not dropped")
	}
	if d := <-drops; d.Text != "some dragged text" {
		t.Errorf("text drop = %+v", d)
	}

	// Released over the bare root window: nobody takes it.
	drag(DragData{Text: "nowhere"}, [][2]int{{20, 20}, {300, 300}})
	if result() {
		t.Error("a drag released over nothing was reported as dropped")
	}

	// Onto the App's own window.
	drag(DragData{Text: "to myself"}, [][2]int{{420, 420}})
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

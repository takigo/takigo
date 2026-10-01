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

//go:build linux || freebsd || openbsd || netbsd

package takigo_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/dialog"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/text"
	"github.com/takigo/takigo/window"
)

// TestPrint prints a text and a canvas widget through stand-ins for the
// CUPS commands and checks what reaches lpr.
func TestPrint(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		args, body := printJob(t, func(app *takigo.App) widget.Widget {
			tw := text.New(app, "t")
			tw.Insert("end", "hello printer")
			return tw
		})
		for _, want := range []string{"-P office", "copies=1", "media=a4", "page-left=70", "-r "} {
			if !strings.Contains(args, want) {
				t.Errorf("lpr arguments %q lack %q", args, want)
			}
		}
		if !strings.HasPrefix(body, "hello printer") {
			t.Errorf("printed text = %q", body)
		}
	})
	t.Run("canvas", func(t *testing.T) {
		args, body := printJob(t, func(app *takigo.App) widget.Widget {
			c := canvas.New(app, "c")
			c.CreateRectangle(10, 10, 50, 50, canvas.FillColor("red"))
			pack.Pack(c)
			return c
		})
		if strings.Contains(args, "page-left") || !strings.Contains(args, "-P office") {
			t.Errorf("canvas lpr arguments = %q", args)
		}
		if !strings.HasPrefix(body, "%!PS-Adobe") {
			t.Errorf("canvas job is not PostScript: %.40q", body)
		}
	})
}

// printJob runs dialog.Print on the widget mk makes, clicks Print and
// returns lpr's arguments and the printed data.
func printJob(t *testing.T, mk func(*takigo.App) widget.Widget) (args, body string) {
	bin := t.TempDir()
	out := filepath.Join(t.TempDir(), "job")
	script := func(name, body string) {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	script("lpstat", `case "$1" in -e) echo office;; -d) echo "system default destination: office";; esac`)
	script("lpoptions", `echo "printer-location='Room 2' printer-state=3"`)
	// The file is the last argument, after -r.
	script("lpr", `for a; do f=$a; done; { echo "$@"; cat "$f"; } > `+out+`.tmp; mv `+out+`.tmp `+out)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	app := testutil.NewTestApp(t)
	w := mk(app)

	var err error
	app.RunOnMain(func() {
		err = dialog.Print(app, w)
	})
	app.RunOnMain(func() {
		var find func(w *window.Window) *window.Window
		find = func(w *window.Window) *window.Window {
			if strings.HasSuffix(w.PathName, ".buttonf.print") {
				return w
			}
			for _, c := range w.Children {
				if f := find(c); f != nil {
					return f
				}
			}
			return nil
		}
		btn := find(app.Root())
		if btn == nil {
			t.Error("no Print button")
			app.Quit()
			return
		}
		app.DoWhenIdle(func() {
			for _, typ := range []event.Type{event.EnterType, event.ButtonPressType, event.ButtonReleaseType} {
				app.Dispatcher().Dispatch(&event.Event{Type: typ, Window: btn.PlatformID, Button: 1, X: 2, Y: 2})
			}
		})
	})
	deadline := time.Now().Add(5 * time.Second)
	var poll func()
	poll = func() {
		if _, e := os.Stat(out); e == nil || time.Now().After(deadline) {
			app.Quit()
			return
		}
		app.After(50*time.Millisecond, poll)
	}
	app.After(100*time.Millisecond, poll)
	app.MainLoop()

	if err != nil {
		t.Fatalf("Print: %v", err)
	}
	job, e := os.ReadFile(out)
	if e != nil {
		t.Fatalf("lpr was not run: %v", e)
	}
	args, body, _ = strings.Cut(string(job), "\n")
	return args, body
}

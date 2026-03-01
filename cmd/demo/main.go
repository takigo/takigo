// Phase 1 demo: Opens a window, draws colored rectangles, responds to
// keyboard/mouse events, closable with 'q' key or window close button.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/window"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Phase 1 Demo"), takigo.Size(600, 400))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	d := app.Display()

	// Colors (pixel values for TrueColor displays).
	red := uint64(0xCC0000)
	green := uint64(0x00AA00)
	blue := uint64(0x3366CC)
	yellow := uint64(0xFFCC00)

	// Draw handler — called on Expose events.
	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return // wait for last expose in sequence
		}
		drawScene(d, root, red, green, blue, yellow)
	})

	// Key handler — 'q' or Escape to quit.
	app.Dispatcher().Bind(root.XWindow, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_q || ev.KeySym == xlib.XK_Escape {
			app.Quit()
			return
		}
		fmt.Printf("Key press: sym=0x%x str=%q\n", ev.KeySym, ev.Str)
	})

	// Mouse button handler.
	app.Dispatcher().Bind(root.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		fmt.Printf("Button %d at (%d, %d)\n", ev.Button, ev.X, ev.Y)
	})

	// Enter/Leave handlers.
	app.Dispatcher().Bind(root.XWindow, event.EnterMask, func(ev *event.Event) {
		fmt.Println("Mouse entered window")
	})
	app.Dispatcher().Bind(root.XWindow, event.LeaveMask, func(ev *event.Event) {
		fmt.Println("Mouse left window")
	})

	fmt.Println("Takigo Phase 1 Demo")
	fmt.Println("Press 'q' or Escape to quit. Click to see coordinates.")
	app.MainLoop()
	fmt.Println("Goodbye!")
}

// drawScene draws colored rectangles on the window.
func drawScene(d *window.Display, root *window.Window, red, green, blue, yellow uint64) {
	xd := d.XDisplay
	gc := root.GC
	drawable := root.Drawable()
	w := root.Width
	h := root.Height

	// Background.
	xd.SetForeground(gc, d.WhitePixel)
	xd.FillRectangle(drawable, gc, 0, 0, uint(w), uint(h))

	// Red rectangle.
	xd.SetForeground(gc, red)
	xd.FillRectangle(drawable, gc, 20, 20, 150, 100)

	// Green rectangle.
	xd.SetForeground(gc, green)
	xd.FillRectangle(drawable, gc, 200, 20, 150, 100)

	// Blue rectangle.
	xd.SetForeground(gc, blue)
	xd.FillRectangle(drawable, gc, 20, 150, 150, 100)

	// Yellow rectangle.
	xd.SetForeground(gc, yellow)
	xd.FillRectangle(drawable, gc, 200, 150, 150, 100)

	// Draw outlines in black.
	xd.SetForeground(gc, d.BlackPixel)
	xd.DrawRectangle(drawable, gc, 20, 20, 150, 100)
	xd.DrawRectangle(drawable, gc, 200, 20, 150, 100)
	xd.DrawRectangle(drawable, gc, 20, 150, 150, 100)
	xd.DrawRectangle(drawable, gc, 200, 150, 150, 100)

	// Draw diagonal lines.
	xd.DrawLine(drawable, gc, 400, 20, 560, 260)
	xd.DrawLine(drawable, gc, 560, 20, 400, 260)

	// Draw some text.
	xd.DrawString(drawable, gc, 420, 300, "Takigo Phase 1")

	xd.Flush()
}

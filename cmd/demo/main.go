// Phase 3 demo: Demonstrates pack, grid, and place geometry managers
// with child windows responsive to resizing.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/geometry/place"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/window"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Phase 3 — Geometry Managers"), takigo.Size(800, 600))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	d := app.Display()

	// Colors.
	colors := color.NewCache(d.XDisplay, d.Screen, d.Colormap)
	bgColor, _ := colors.Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	red, _ := colors.Get("#cc4444")
	green, _ := colors.Get("#44aa44")
	blue, _ := colors.Get("#4466cc")
	yellow, _ := colors.Get("#ccaa00")
	cyan, _ := colors.Get("#44aaaa")
	purple, _ := colors.Get("#884488")

	// Font.
	fontReg := font.NewRegistry(d.XDisplay, d.Screen, d.Visual, d.Colormap)
	defer fontReg.Close()
	defFont, _ := fontReg.Get(font.TkDefaultFont)
	xftFont := defFont.(*font.XftFont)

	// Create child windows for pack demo.
	packTop := window.NewChildWindow(root, "ptop", 0, 0, 100, 40)
	packTop.BackgroundPixel = red.Pixel
	window.MakeWindowExist(packTop)

	packBottom := window.NewChildWindow(root, "pbot", 0, 0, 100, 40)
	packBottom.BackgroundPixel = green.Pixel
	window.MakeWindowExist(packBottom)

	packLeft := window.NewChildWindow(root, "pleft", 0, 0, 80, 40)
	packLeft.BackgroundPixel = blue.Pixel
	window.MakeWindowExist(packLeft)

	packRight := window.NewChildWindow(root, "pright", 0, 0, 80, 40)
	packRight.BackgroundPixel = yellow.Pixel
	window.MakeWindowExist(packRight)

	packCenter := window.NewChildWindow(root, "pcenter", 0, 0, 100, 40)
	packCenter.BackgroundPixel = cyan.Pixel
	window.MakeWindowExist(packCenter)

	// Pack children.
	pack.Pack(packTop, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))
	pack.Pack(packBottom, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(2))
	pack.Pack(packLeft, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX(2))
	pack.Pack(packRight, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY), pack.PadX(2))
	pack.Pack(packCenter, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Handle ConfigureNotify for resize.
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
			redrawAll(d, root, xftFont, bgColor,
				packTop, packBottom, packLeft, packRight, packCenter,
				red, green, blue, yellow, cyan)
		}
	})

	// Draw handler.
	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		redrawAll(d, root, xftFont, bgColor,
			packTop, packBottom, packLeft, packRight, packCenter,
			red, green, blue, yellow, cyan)
	})

	// Per-child expose handlers.
	childWindows := []*window.Window{packTop, packBottom, packLeft, packRight, packCenter}
	childColors := []*color.Color{red, green, blue, yellow, cyan}
	childLabels := []string{"Top (pack)", "Bottom (pack)", "Left (pack)", "Right (pack)", "Center (expand)"}

	for i, child := range childWindows {
		col := childColors[i]
		label := childLabels[i]
		w := child
		app.Dispatcher().Bind(w.XWindow, event.ExposureMask, func(ev *event.Event) {
			if ev.ExposeCount > 0 {
				return
			}
			drawChild(d, w, xftFont, col, label)
		})
	}

	// Key handler.
	app.Dispatcher().Bind(root.XWindow, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_q || ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	// Suppress unused import errors.
	_ = grid.Row
	_ = place.X
	_ = purple
	_ = draw.NewBorder
	_ = option.ReliefFlat

	fmt.Println("Takigo Phase 3 Demo — Geometry Managers")
	fmt.Println("Resize the window to see pack layout respond. Press 'q' to quit.")
	app.MainLoop()
	fmt.Println("Goodbye!")
}

func redrawAll(d *window.Display, root *window.Window, xftFont *font.XftFont,
	bgColor *color.Color,
	packTop, packBottom, packLeft, packRight, packCenter *window.Window,
	red, green, blue, yellow, cyan *color.Color) {

	xd := d.XDisplay
	gc := root.GC

	// Fill root background.
	xd.SetForeground(gc, bgColor.Pixel)
	xd.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
	xd.Flush()
}

func drawChild(d *window.Display, w *window.Window, xftFont *font.XftFont,
	col *color.Color, label string) {
	xd := d.XDisplay
	gc := w.GC

	// Fill background.
	xd.SetForeground(gc, col.Pixel)
	xd.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw label.
	xftFont.DrawString(w.Drawable(), 5, 20, label, 0xFFFFFF, 0xFFFF, 0xFFFF, 0xFFFF)

	// Draw size info.
	sizeStr := fmt.Sprintf("%dx%d", w.Width, w.Height)
	xftFont.DrawString(w.Drawable(), 5, 35, sizeStr, 0xFFFFFF, 0xFFFF, 0xFFFF, 0xFFFF)

	xd.Flush()
}

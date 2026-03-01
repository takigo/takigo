// Phase 2 demo: Demonstrates text with named fonts, colored rectangles with
// 3D relief borders, and the functional options framework.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/window"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Phase 2 Demo"), takigo.Size(700, 500))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	d := app.Display()

	// Create color cache and parse some colors.
	colors := color.NewCache(d.XDisplay, d.Screen, d.Colormap)
	bgColor, _ := colors.Get("#d9d9d9")  // Tk default background
	red, _ := colors.Get("firebrick")
	blue, _ := colors.Get("steel blue")
	green, _ := colors.Get("forest green")
	gold, _ := colors.Get("gold")

	// Set window background to Tk-like gray.
	if bgColor != nil {
		root.BackgroundPixel = bgColor.Pixel
	}

	// Create borders for 3D relief.
	bgBorder := draw.NewBorderFromPixel(bgColor.Pixel)
	redBorder := draw.NewBorder(red.Red, red.Green, red.Blue)
	blueBorder := draw.NewBorder(blue.Red, blue.Green, blue.Blue)

	// Create font registry and open fonts.
	fontReg := font.NewRegistry(d.XDisplay, d.Screen, d.Visual, d.Colormap)
	defer fontReg.Close()

	defaultFont, err := fontReg.Get(font.TkDefaultFont)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Font error: %v\n", err)
		os.Exit(1)
	}

	headingFont, err := fontReg.Get(font.TkHeadingFont)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Font error: %v\n", err)
		os.Exit(1)
	}

	fixedFont, err := fontReg.Get(font.TkFixedFont)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Font error: %v\n", err)
		os.Exit(1)
	}

	// Draw handler.
	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		drawScene(d, root, bgColor, bgBorder, redBorder, blueBorder,
			red, blue, green, gold,
			defaultFont.(*font.XftFont), headingFont.(*font.XftFont), fixedFont.(*font.XftFont))
	})

	// Key handler.
	app.Dispatcher().Bind(root.XWindow, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_q || ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	fmt.Println("Takigo Phase 2 Demo — Colors, Fonts, 3D Relief")
	fmt.Println("Press 'q' or Escape to quit.")
	app.MainLoop()
	fmt.Println("Goodbye!")
}

func drawScene(d *window.Display, root *window.Window,
	bgColor *color.Color, bgBorder, redBorder, blueBorder *draw.Border,
	red, blue, green, gold *color.Color,
	defaultFont, headingFont, fixedFont *font.XftFont) {

	xd := d.XDisplay
	gc := root.GC
	drawable := root.Drawable()
	w := root.Width
	h := root.Height

	// Fill background.
	xd.SetForeground(gc, bgColor.Pixel)
	xd.FillRectangle(drawable, gc, 0, 0, uint(w), uint(h))

	// Title text with heading font.
	headingFont.DrawString(drawable, 20, 35, "Takigo Phase 2: Colors, Fonts & 3D Relief",
		d.BlackPixel, 0, 0, 0)

	// 3D Relief demo — show all relief types.
	reliefs := []struct {
		name   string
		relief option.Relief
	}{
		{"Raised", option.ReliefRaised},
		{"Sunken", option.ReliefSunken},
		{"Groove", option.ReliefGroove},
		{"Ridge", option.ReliefRidge},
		{"Solid", option.ReliefSolid},
		{"Flat", option.ReliefFlat},
	}

	y := 60
	for i, r := range reliefs {
		x := 20 + i*110
		draw.Fill3DRectangle(xd, drawable, gc, bgBorder, x, y, 100, 60, 3, r.relief)
		// Label with default font.
		defaultFont.DrawString(drawable, x+10, y+35, r.name,
			d.BlackPixel, 0, 0, 0)
	}

	// Colored rectangles with 3D relief.
	draw.Fill3DRectangle(xd, drawable, gc, redBorder, 20, 150, 200, 80, 3, option.ReliefRaised)
	defaultFont.DrawString(drawable, 40, 195, "Red Raised",
		0xFFFFFF, 0xFFFF, 0xFFFF, 0xFFFF)

	draw.Fill3DRectangle(xd, drawable, gc, blueBorder, 240, 150, 200, 80, 3, option.ReliefSunken)
	defaultFont.DrawString(drawable, 260, 195, "Blue Sunken",
		0xFFFFFF, 0xFFFF, 0xFFFF, 0xFFFF)

	draw.Fill3DRectangle(xd, drawable, gc, bgBorder, 460, 150, 200, 80, 4, option.ReliefRidge)
	defaultFont.DrawString(drawable, 480, 195, "Gray Ridge",
		d.BlackPixel, 0, 0, 0)

	// Named colors demo.
	namedColors := []struct {
		name string
		col  *color.Color
	}{
		{"firebrick", red},
		{"steel blue", blue},
		{"forest green", green},
		{"gold", gold},
	}

	y = 260
	for i, nc := range namedColors {
		x := 20 + i*170
		xd.SetForeground(gc, nc.col.Pixel)
		xd.FillRectangle(drawable, gc, x, y, 150, 40)
		xd.SetForeground(gc, d.BlackPixel)
		xd.DrawRectangle(drawable, gc, x, y, 150, 40)
		defaultFont.DrawString(drawable, x+5, y+25, nc.name,
			d.BlackPixel, 0, 0, 0)
	}

	// Font demo.
	y = 330
	headingFont.DrawString(drawable, 20, y, "Heading Font (TkHeadingFont — sans-serif bold 12)",
		d.BlackPixel, 0, 0, 0)

	y += 30
	defaultFont.DrawString(drawable, 20, y, "Default Font (TkDefaultFont — sans-serif 10)",
		d.BlackPixel, 0, 0, 0)

	y += 25
	fixedFont.DrawString(drawable, 20, y, "Fixed Font (TkFixedFont — monospace 10)",
		d.BlackPixel, 0, 0, 0)

	// Font metrics display.
	y += 30
	m := headingFont.Metrics()
	metricsStr := fmt.Sprintf("Heading metrics: ascent=%d descent=%d maxWidth=%d fixed=%v",
		m.Ascent, m.Descent, m.MaxWidth, m.Fixed)
	fixedFont.DrawString(drawable, 20, y, metricsStr,
		d.BlackPixel, 0, 0, 0)

	// Text measurement demo.
	y += 25
	testStr := "Hello, Takigo!"
	textWidth := headingFont.MeasureString(testStr)
	measureStr := fmt.Sprintf("MeasureString(%q) = %d pixels", testStr, textWidth)
	fixedFont.DrawString(drawable, 20, y, measureStr,
		d.BlackPixel, 0, 0, 0)

	xd.Flush()
}

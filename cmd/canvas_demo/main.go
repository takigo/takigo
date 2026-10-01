// Phase 10 demo: Canvas widget with various item types.
package main

import (
	"fmt"
	goimage "image"
	"image/color"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scrollbar"
)

// generateTestImage creates a 48x48 RGBA image with a gradient.
func generateTestImage() *goimage.RGBA {
	img := goimage.NewRGBA(goimage.Rect(0, 0, 48, 48))
	for y := range 48 {
		for x := range 48 {
			r := uint8(x * 255 / 48)
			g := uint8(y * 255 / 48)
			b := uint8(128)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Phase 10 — Canvas Widget"), takigo.Size(800, 650))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.SetBackgroundPixel(bgColor.Pixel)

	// Status label at bottom.
	statusLabel := label.New(app, "status",
		label.Text("Phase 10: Canvas Widget. Hover items for events. Esc to quit."),
		label.Background("#e8e8e8"),
		label.Anchor(option.AnchorW),
		label.PadX(5),
		label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Create canvas with scrollbars.
	cv := canvas.New(app, "canvas",
		canvas.Width(760),
		canvas.Height(550),
		canvas.Background("white"),
		canvas.BorderWidthOpt(2),
		canvas.ReliefOpt(option.ReliefSunken),
		canvas.ScrollRegion(0, 0, 1200, 1000),
	)

	// Horizontal scrollbar.
	xscroll := scrollbar.New(app, "xscroll",
		scrollbar.OrientOpt(scrollbar.Horizontal),
		scrollbar.WidthOpt(14),
		scrollbar.CommandOpt(widget.ScrollX(cv)),
	)
	cv.XScrollCmd = func(first, last float64) {
		xscroll.Set(first, last)
	}

	// Vertical scrollbar.
	yscroll := scrollbar.New(app, "yscroll",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.WidthOpt(14),
		scrollbar.CommandOpt(widget.ScrollY(cv)),
	)
	cv.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	// Pack scrollbars and canvas.
	pack.Pack(xscroll, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))
	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(cv, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Initialize scrollbar positions.
	xf, xl := cv.XVisibleRange()
	xscroll.Set(xf, xl)
	yf, yl := cv.YVisibleRange()
	yscroll.Set(yf, yl)

	// --- Create canvas items ---

	// Rectangles.
	cv.CreateRectangle(20, 20, 150, 100,
		canvas.FillColor("#4a90d9"),
		canvas.OutlineColor("#2c5ea0"),
		canvas.OutlineWidth(2),
		canvas.Tags("shapes", "rect1"),
	)
	cv.CreateRectangle(170, 20, 300, 100,
		canvas.FillColor("#e8c840"),
		canvas.OutlineColor("#b89830"),
		canvas.OutlineWidth(2),
		canvas.Dash(6, 3),
		canvas.Tags("shapes", "rect2"),
	)

	// Ovals.
	cv.CreateOval(320, 20, 480, 100,
		canvas.FillColor("#50c878"),
		canvas.OutlineColor("#308050"),
		canvas.OutlineWidth(2),
		canvas.Tags("shapes", "oval1"),
	)
	cv.CreateOval(500, 20, 620, 100,
		canvas.FillColor("#e85050"),
		canvas.Tags("shapes", "oval2"),
	)

	// Lines.
	cv.CreateLine([]float64{20, 130, 100, 200, 180, 130, 260, 200},
		canvas.OutlineColor("#333333"),
		canvas.OutlineWidth(3),
		canvas.Tags("lines"),
	)
	cv.CreateLine([]float64{300, 130, 380, 200, 460, 130, 540, 200},
		canvas.OutlineColor("#8040c0"),
		canvas.OutlineWidth(2),
		canvas.Smooth(true),
		canvas.SplineSteps(20),
		canvas.Tags("lines", "smooth"),
	)

	// Line with arrows.
	cv.CreateLine([]float64{580, 130, 700, 200},
		canvas.OutlineColor("#c04040"),
		canvas.OutlineWidth(2),
		canvas.Arrow(canvas.ArrowBoth),
		canvas.ArrowShape(10, 12, 4),
		canvas.Tags("lines", "arrows"),
	)

	// Polygons.
	// Triangle.
	cv.CreatePolygon([]float64{50, 240, 130, 320, 50, 320},
		canvas.FillColor("#ff8c00"),
		canvas.OutlineColor("#c06800"),
		canvas.OutlineWidth(2),
		canvas.Tags("polygons"),
	)
	// Star.
	cv.CreatePolygon([]float64{
		230, 240, 250, 290, 300, 290, 260, 320,
		275, 370, 230, 340, 185, 370, 200, 320,
		160, 290, 210, 290,
	},
		canvas.FillColor("#ffd700"),
		canvas.OutlineColor("#b8a000"),
		canvas.OutlineWidth(2),
		canvas.Tags("polygons", "star"),
	)

	// Arcs.
	cv.CreateArc(350, 240, 480, 370,
		canvas.StartAngle(30),
		canvas.Extent(120),
		canvas.ArcStyleOpt(canvas.ArcStylePieslice),
		canvas.FillColor("#87ceeb"),
		canvas.OutlineColor("#4682b4"),
		canvas.OutlineWidth(2),
		canvas.Tags("arcs"),
	)
	cv.CreateArc(500, 240, 630, 370,
		canvas.StartAngle(0),
		canvas.Extent(270),
		canvas.ArcStyleOpt(canvas.ArcStyleChord),
		canvas.FillColor("#dda0dd"),
		canvas.OutlineColor("#8b668b"),
		canvas.OutlineWidth(2),
		canvas.Tags("arcs"),
	)
	cv.CreateArc(660, 240, 780, 370,
		canvas.StartAngle(45),
		canvas.Extent(180),
		canvas.ArcStyleOpt(canvas.ArcStyleArc),
		canvas.OutlineColor("#2f4f4f"),
		canvas.OutlineWidth(3),
		canvas.Tags("arcs"),
	)

	// Text.
	cv.CreateText(60, 400,
		canvas.TextOpt("Hello, Canvas!"),
		canvas.TextColor("#000080"),
		canvas.AnchorOpt(option.AnchorNW),
		canvas.Tags("text"),
	)
	cv.CreateText(300, 400,
		canvas.TextOpt("Centered text"),
		canvas.TextColor("#800000"),
		canvas.AnchorOpt(option.AnchorN),
		canvas.Tags("text"),
	)

	// Image.
	testRGBA := generateTestImage()
	testPhoto := tkimage.NewPhoto("canvastest", testRGBA)
	app.ImageRegistry().Register(testPhoto)

	cv.CreateImage(550, 420,
		canvas.ImageOpt(testPhoto),
		canvas.AnchorOpt(option.AnchorNW),
		canvas.Tags("images"),
	)

	// Items far away (to test scrolling).
	cv.CreateRectangle(800, 600, 1100, 900,
		canvas.FillColor("#e0e0ff"),
		canvas.OutlineColor("#8080c0"),
		canvas.OutlineWidth(3),
		canvas.Tags("far"),
	)
	cv.CreateText(950, 750,
		canvas.TextOpt("Scroll to see me!"),
		canvas.TextColor("#4040a0"),
		canvas.Tags("far", "text"),
	)

	// --- Item event bindings ---
	cv.BindItem("shapes", event.EnterMask, func(ev *event.Event) {
		statusLabel.Configure(label.Text("Hovering over a shape"))
	})
	cv.BindItem("shapes", event.LeaveMask, func(ev *event.Event) {
		statusLabel.Configure(label.Text("Phase 10: Canvas Widget. Hover items for events. Esc to quit."))
	})
	cv.BindItem("star", event.ButtonPressMask, func(ev *event.Event) {
		fmt.Println("Star clicked! Moving it by (5, 5)...")
		cv.Move("star", 5, 5)
	})

	// --- Root event handlers ---
	app.Dispatcher().Bind(root.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	app.Dispatcher().Bind(root.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.Server
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_Escape {
			app.Quit()
		}
	})

	fmt.Println("Takigo Phase 10 Demo — Canvas Widget")
	fmt.Println("Click the star to move it. Scroll with scrollbars. Esc to quit.")
	app.Run()
	fmt.Println("Goodbye!")

	// Keep references alive.
	_ = statusLabel
	_ = cv
	_ = xscroll
	_ = yscroll

	// Suppress unused import.
	_ = widget.CompoundNone
}

// Demo: Canvas widget with examples of the various canvas item types.
// Ported from Tk's items.tcl demo.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	tkimage "github.com/takigo/takigo/image"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/entry"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/scale"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Canvas Item Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("Items"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(5)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window contains a canvas widget with examples of the various kinds of items supported by canvases.  The following operations are supported:\n  Left-Button drag:\tmoves item under pointer.\n  Middle-Button drag:\trepositions view.\n  Right-Button drag:\tstrokes out area.\n  Ctrl+f:\t\tprints items under area."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Outer grid frame for canvas + scrollbars.
	gf := frame.New(f, "frame")
	pack.Pack(gf, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Canvas with scroll region covering all nine sections.
	c := canvas.New(gf, "c",
		canvas.Width(screenunit.Cm(15).Pixels()),
		canvas.Height(screenunit.Cm(10).Pixels()),
		canvas.ScrollRegion(0, 0, screenunit.Cm(30).Pixels(), screenunit.Cm(24).Pixels()),
		canvas.ReliefOpt(option.ReliefSunken),
		canvas.BorderWidthOpt(2),
	)

	yscroll := ttk.NewScrollbar(gf, "vscroll",
		ttk.ScrollbarOrient(ttk.Vertical),
		ttk.ScrollbarCommand(widget.ScrollY(c)),
	)
	xscroll := ttk.NewScrollbar(gf, "hscroll",
		ttk.ScrollbarOrient(ttk.Horizontal),
		ttk.ScrollbarCommand(widget.ScrollX(c)),
	)
	c.Configure(
		canvas.YScrollCommand(func(first, last float64) { yscroll.Set(first, last) }),
		canvas.XScrollCommand(func(first, last float64) { xscroll.Set(first, last) }),
	)

	// Grid layout: canvas(0,0), yscroll(0,1), xscroll(1,0).
	grid.Grid(c, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NSEW))
	grid.Grid(xscroll, grid.Row(1), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.RowConfigure(gf, 0, grid.Weight(1), grid.MinSize(0))
	grid.ColumnConfigure(gf, 0, grid.Weight(1), grid.MinSize(0))

	// Helper: convert centimeters to canvas pixel coordinate.
	pxPerCm := screenunit.Cm(1).Float()
	p := func(cm float64) float64 { return cm * pxPerCm }

	// Helper: convert a distance to int pixels.
	px := screenunit.Distance.Pixels

	// Colors matching Tk's color depth >1 defaults.
	const blue = "#009ACD" // DeepSkyBlue3
	const red = "red"
	const green = "SeaGreen3"

	// ---- Structural 3×3 grid ----
	c.CreateRectangle(0, 0, p(30), p(24), canvas.OutlineColor("black"), canvas.OutlineWidth(px(screenunit.Pt(1.5))))
	c.CreateLine([]float64{0, p(8), p(30), p(8)}, canvas.OutlineColor("black"), canvas.OutlineWidth(px(screenunit.Pt(1.5))))
	c.CreateLine([]float64{0, p(16), p(30), p(16)}, canvas.OutlineColor("black"), canvas.OutlineWidth(px(screenunit.Pt(1.5))))
	c.CreateLine([]float64{p(10), 0, p(10), p(24)}, canvas.OutlineColor("black"), canvas.OutlineWidth(px(screenunit.Pt(1.5))))
	c.CreateLine([]float64{p(20), 0, p(20), p(24)}, canvas.OutlineColor("black"), canvas.OutlineWidth(px(screenunit.Pt(1.5))))

	// Track original colors per item ID for hover restore.
	type itemColors struct {
		fill    string
		outline string
		isText  bool
	}
	origColors := map[canvas.ItemID]itemColors{}

	record := func(id canvas.ItemID, fill, outline string) {
		origColors[id] = itemColors{fill: fill, outline: outline}
	}
	recordText := func(id canvas.ItemID, textColor string) {
		origColors[id] = itemColors{fill: textColor, isText: true}
	}

	// ---- Section 1: Lines (col 0, row 0) ----
	c.CreateText(p(5), p(0.2), canvas.TextOpt("Lines"), canvas.AnchorOpt(option.AnchorN))

	// Bold "Z" shape in blue.
	record(c.CreateLine([]float64{p(1), p(1), p(3), p(1), p(1), p(4), p(3), p(4)},
		canvas.OutlineColor(blue), canvas.OutlineWidth(px(screenunit.Mm(2))),
		canvas.CapStyleOpt(platform.CapButt), canvas.JoinStyleOpt(platform.JoinMiter),
		canvas.Tags("item")), "", blue)

	// Single arrow (down).
	record(c.CreateLine([]float64{p(4.67), p(1), p(4.67), p(4)},
		canvas.Arrow(canvas.ArrowLast),
		canvas.Tags("item")), "", "black")

	// Double arrow (up+down).
	record(c.CreateLine([]float64{p(6.33), p(1), p(6.33), p(4)},
		canvas.Arrow(canvas.ArrowBoth),
		canvas.Tags("item")), "", "black")

	// Nested rectangles (spiral line) in red.
	record(c.CreateLine([]float64{
		p(5), p(6), p(9), p(6), p(9), p(1), p(8), p(1), p(8), p(4.8),
		p(8.8), p(4.8), p(8.8), p(1.2), p(8.2), p(1.2), p(8.2), p(4.6),
		p(8.6), p(4.6), p(8.6), p(1.4), p(8.4), p(1.4), p(8.4), p(4.4),
	}, canvas.OutlineColor(red), canvas.OutlineWidth(px(screenunit.Pt(2.25))),
		canvas.Tags("item")), "", red)

	gray25 := "@" + findImage("gray25.xbm")

	record(c.CreateLine([]float64{p(1), p(5), p(7), p(5), p(7), p(7), p(9), p(7)},
		canvas.OutlineWidth(px(screenunit.Cm(0.5))),
		canvas.Stipple(gray25),
		canvas.Arrow(canvas.ArrowBoth),
		canvas.ArrowShape(15, 15, 7),
		canvas.Tags("item")), "", "black")

	// Wavy line with round caps/joins.
	record(c.CreateLine([]float64{p(1), p(7), p(1.75), p(5.8), p(2.5), p(7), p(3.25), p(5.8), p(4), p(7)},
		canvas.OutlineWidth(px(screenunit.Cm(0.5))),
		canvas.CapStyleOpt(platform.CapRound),
		canvas.JoinStyleOpt(platform.JoinRound),
		canvas.Tags("item")), "", "black")

	// ---- Section 2: Curves (col 1, row 0) ----
	c.CreateText(p(15), p(0.2), canvas.TextOpt("Curves (smoothed lines)"), canvas.AnchorOpt(option.AnchorN))

	// Bell curve in blue.
	record(c.CreateLine([]float64{p(11), p(4), p(11.5), p(1), p(13.5), p(1), p(14), p(4)},
		canvas.OutlineColor(blue), canvas.Smooth(true),
		canvas.Tags("item")), "", blue)

	// Crossing arrows smooth curve.
	record(c.CreateLine([]float64{p(15.5), p(1), p(19.5), p(1.5), p(15.5), p(4.5), p(19.5), p(4)},
		canvas.Smooth(true),
		canvas.Arrow(canvas.ArrowBoth),
		canvas.OutlineWidth(px(screenunit.Pt(2.25))),
		canvas.Tags("item")), "", "black")

	record(c.CreateLine([]float64{
		p(12), p(6), p(13.5), p(4.5), p(16.5), p(7.5), p(18), p(6),
		p(16.5), p(4.5), p(13.5), p(7.5), p(12), p(6),
	}, canvas.OutlineColor(red), canvas.Smooth(true),
		canvas.OutlineWidth(px(screenunit.Mm(3))),
		canvas.CapStyleOpt(platform.CapRound),
		canvas.Stipple(gray25),
		canvas.Tags("item")), "", red)

	// ---- Section 3: Polygons (col 2, row 0) ----
	c.CreateText(p(25), p(0.2), canvas.TextOpt("Polygons"), canvas.AnchorOpt(option.AnchorN))

	// 8-point star in green, no outline.
	record(c.CreatePolygon([]float64{
		p(21), p(1.0), p(22.5), p(1.75), p(24), p(1.0), p(23.25), p(2.5),
		p(24), p(4.0), p(22.5), p(3.25), p(21), p(4.0), p(21.75), p(2.5),
	}, canvas.FillColor(green), canvas.OutlineNone(), canvas.OutlineWidth(px(screenunit.Pt(3))),
		canvas.Tags("item")), green, "")

	// Smooth M-wave polygon in red, no outline.
	record(c.CreatePolygon([]float64{
		p(25), p(4), p(25), p(4), p(25), p(1), p(26), p(1), p(27), p(4),
		p(28), p(1), p(29), p(1), p(29), p(4), p(29), p(4),
	}, canvas.FillColor(red), canvas.OutlineNone(), canvas.Smooth(true),
		canvas.Tags("item")), red, "")

	record(c.CreatePolygon([]float64{
		p(22), p(4.5), p(25), p(4.5), p(25), p(6.75), p(28), p(6.75),
		p(28), p(5.25), p(24), p(5.25), p(24), p(6.0), p(26), p(6), p(26), p(7.5), p(22), p(7.5),
	}, canvas.FillColor(blue), canvas.OutlineNone(), canvas.Stipple(gray25),
		canvas.Tags("item")), blue, "")

	// ---- Section 4: Rectangles (col 0, row 1) ----
	c.CreateText(p(5), p(8.2), canvas.TextOpt("Rectangles"), canvas.AnchorOpt(option.AnchorN))

	// Red outline only rectangle.
	record(c.CreateRectangle(p(1), p(9.5), p(4), p(12.5),
		canvas.FillNone(), canvas.OutlineColor(red), canvas.OutlineWidth(px(screenunit.Mm(3))),
		canvas.Tags("item")), "", red)

	// Green filled rectangle (default black outline).
	record(c.CreateRectangle(p(0.5), p(13.5), p(4.5), p(15.5),
		canvas.FillColor(green),
		canvas.Tags("item")), green, "black")

	record(c.CreateRectangle(p(6), p(10), p(9), p(15),
		canvas.FillColor(blue), canvas.OutlineNone(), canvas.Stipple(gray25),
		canvas.Tags("item")), blue, "")

	// ---- Section 5: Ovals (col 1, row 1) ----
	c.CreateText(p(15), p(8.2), canvas.TextOpt("Ovals"), canvas.AnchorOpt(option.AnchorN))

	// Red outline only oval.
	record(c.CreateOval(p(11), p(9.5), p(14), p(12.5),
		canvas.FillNone(), canvas.OutlineColor(red), canvas.OutlineWidth(px(screenunit.Mm(3))),
		canvas.Tags("item")), "", red)

	// Green filled oval (default black outline).
	record(c.CreateOval(p(10.5), p(13.5), p(14.5), p(15.5),
		canvas.FillColor(green),
		canvas.Tags("item")), green, "black")

	record(c.CreateOval(p(16), p(10), p(19), p(15),
		canvas.FillColor(blue), canvas.OutlineNone(), canvas.Stipple(gray25),
		canvas.Tags("item")), blue, "")

	// ---- Section 6: Text (col 2, row 1) ----
	c.CreateText(p(25), p(8.2), canvas.TextOpt("Text"), canvas.AnchorOpt(option.AnchorN))

	// Anchor point marker.
	c.CreateRectangle(p(22.4), p(8.9), p(22.6), p(9.1), canvas.OutlineColor("black"))

	// Word-wrapped text, anchor N, justified left.
	recordText(c.CreateText(p(22.5), p(9),
		canvas.TextOpt("A short string of text, word-wrapped, justified left, and anchored north (at the top).  The rectangles show the anchor points for each piece of text."),
		canvas.FontOpt("Helvetica 12"),
		canvas.WidthOpt(px(screenunit.Cm(4))),
		canvas.AnchorOpt(option.AnchorN),
		canvas.JustifyOpt(option.JustifyLeft),
		canvas.Tags("item")), "black")

	// Anchor point marker.
	c.CreateRectangle(p(25.4), p(10.9), p(25.6), p(11.1), canvas.OutlineColor("black"))

	// Multi-line centered text, anchor W, blue.
	recordText(c.CreateText(p(25.5), p(11),
		canvas.TextOpt("Several lines,\n each centered\nindividually,\nand all anchored\nat the left edge."),
		canvas.FontOpt("Helvetica 12"),
		canvas.TextColor(blue),
		canvas.AnchorOpt(option.AnchorW),
		canvas.JustifyOpt(option.JustifyCenter),
		canvas.Tags("item")), blue)

	// Anchor point marker.
	c.CreateRectangle(p(24.9), p(13.9), p(25.1), p(14.1), canvas.OutlineColor("black"))

	// Angled text in red.
	recordText(c.CreateText(p(25), p(14),
		canvas.TextOpt("Angled characters"),
		canvas.FontOpt("Helvetica 24 bold"),
		canvas.TextColor(red),
		canvas.TextAngle(15),
		canvas.AnchorOpt(option.AnchorCenter),
		canvas.Tags("item")), red)

	// ---- Section 7: Arcs (col 0, row 2) ----
	c.CreateText(p(5), p(16.2), canvas.TextOpt("Arcs"), canvas.AnchorOpt(option.AnchorN))

	// Green pieslice.
	record(c.CreateArc(p(0.5), p(17), p(7), p(20),
		canvas.FillColor(green), canvas.OutlineColor("black"),
		canvas.StartAngle(45), canvas.Extent(270),
		canvas.ArcStyleOpt(canvas.ArcStylePieslice),
		canvas.Tags("item")), green, "black")

	record(c.CreateArc(p(6.5), p(17), p(9.5), p(20),
		canvas.OutlineColor(blue), canvas.OutlineWidth(px(screenunit.Mm(4))), canvas.OutlineStipple(gray25),
		canvas.StartAngle(-135), canvas.Extent(270),
		canvas.ArcStyleOpt(canvas.ArcStyleArc),
		canvas.Tags("item")), "", blue)

	// Red pieslice, no fill.
	record(c.CreateArc(p(0.5), p(20), p(9.5), p(24),
		canvas.FillNone(), canvas.OutlineColor(red), canvas.OutlineWidth(px(screenunit.Mm(4))),
		canvas.StartAngle(225), canvas.Extent(-90),
		canvas.ArcStyleOpt(canvas.ArcStylePieslice),
		canvas.Tags("item")), "", red)

	// Blue chord, no outline.
	record(c.CreateArc(p(5.5), p(20.5), p(9.5), p(23.5),
		canvas.FillColor(blue), canvas.OutlineNone(), canvas.OutlineWidth(px(screenunit.Mm(4))),
		canvas.StartAngle(45), canvas.Extent(270),
		canvas.ArcStyleOpt(canvas.ArcStyleChord),
		canvas.Tags("item")), blue, "")

	// ---- Section 8: Bitmaps and Images (col 1, row 2) ----
	c.CreateText(p(15), p(16.2), canvas.TextOpt("Bitmaps and Images"), canvas.AnchorOpt(option.AnchorN))

	// Photo image (ouster.png).
	if imgPath := findImage("ouster.png"); imgPath != "" {
		if photo, err := tkimage.NewPhotoFromFile("items_ouster", imgPath); err == nil {
			app.ImageRegistry().Register(photo)
			c.CreateImage(p(13), p(20),
				canvas.ImageOpt(photo),
				canvas.AnchorOpt(option.AnchorCenter),
				canvas.Tags("item"))
		}
	}

	// XBM bitmaps.
	for _, bname := range []struct{ file, name string }{
		{"noletter.xbm", "noletter"},
		{"letters.xbm", "letters"},
	} {
		if bPath := findImage(bname.file); bPath != "" {
			if xbm, err := canvas.ParseXBMFile(bPath); err == nil {
				_ = bname.name
				yPos := p(18.5)
				if bname.file == "letters.xbm" {
					yPos = p(21.5)
				}
				c.CreateBitmap(p(17), yPos, xbm,
					canvas.AnchorOpt(option.AnchorCenter),
					canvas.BitmapForeground(0, 0, 0),
					canvas.Tags("item"))
			}
		}
	}

	// ---- Section 9: Windows (col 2, row 2) ----
	c.CreateText(p(25), p(16.2), canvas.TextOpt("Windows"), canvas.AnchorOpt(option.AnchorN))

	// Label text "Button:".
	c.CreateText(p(21), p(17.9), canvas.TextOpt("Button:"), canvas.AnchorOpt(option.AnchorSW))

	// Embedded button.
	var btnPressTextID canvas.ItemID
	btn := button.New(c, "win_button",
		button.Text("Press Me"),
		button.Command(func() {
			if btnPressTextID != 0 {
				c.Delete(btnPressTextID)
			}
			btnPressTextID = c.CreateText(p(25), p(18.1),
				canvas.TextOpt("Oooohhh!!"),
				canvas.TextColor(red),
				canvas.AnchorOpt(option.AnchorN))
			app.After(500, func() {
				c.Delete(btnPressTextID)
				btnPressTextID = 0
			})
		}),
	)
	c.CreateWindow(p(21), p(18), btn.Win, canvas.AnchorOpt(option.AnchorNW), canvas.Tags("item"))

	// Label text "Entry:".
	c.CreateText(p(21), p(20.9), canvas.TextOpt("Entry:"), canvas.AnchorOpt(option.AnchorSW))

	// Embedded entry.
	ent := entry.New(c, "win_entry", entry.Width(20))
	ent.Insert(0, "Edit this text")
	c.CreateWindow(p(21), p(21), ent.Win, canvas.AnchorOpt(option.AnchorNW), canvas.Tags("item"))

	// Label text "Scale:".
	c.CreateText(p(28.5), p(17.4), canvas.TextOpt("Scale:"), canvas.AnchorOpt(option.AnchorS))

	// Embedded scale.
	sc := scale.New(c, "win_scale",
		scale.FromOpt(0), scale.ToOpt(100),
		scale.LengthOpt(screenunit.Cm(6).Pixels()),
		scale.SliderLengthOpt(screenunit.Cm(0.4).Pixels()),
		scale.WidthOpt(screenunit.Cm(0.5).Pixels()),
		scale.TickIntervalOpt(0),
	)
	c.CreateWindow(p(28.5), p(17.5), sc.Win, canvas.AnchorOpt(option.AnchorN), canvas.Tags("item"))

	// ---- Event bindings ----

	var highlightedID canvas.ItemID

	c.BindItem("item", event.EnterMask, func(ev *event.Event) {
		highlightedID = 0
		ids := c.FindWithTag("current")
		if len(ids) == 0 {
			return
		}
		id := ids[0]
		colors, ok := origColors[id]
		if !ok {
			return
		}
		highlightedID = id
		if colors.isText {
			c.ItemConfigure("current", canvas.TextColor("#5CACE5")) // SteelBlue2
		} else if colors.fill != "" {
			c.ItemConfigure("current", canvas.FillColor("#5CACE5")) // SteelBlue2
		} else if colors.outline != "" {
			c.ItemConfigure("current", canvas.OutlineColor("#5CACE5")) // SteelBlue2
		}
	})

	c.BindItem("item", event.LeaveMask, func(ev *event.Event) {
		if highlightedID == 0 {
			return
		}
		colors, ok := origColors[highlightedID]
		if !ok {
			return
		}
		if colors.isText {
			c.ItemConfigure(highlightedID, canvas.TextColor(colors.fill))
		} else {
			if colors.fill != "" {
				c.ItemConfigure(highlightedID, canvas.FillColor(colors.fill))
			}
			if colors.outline != "" {
				c.ItemConfigure(highlightedID, canvas.OutlineColor(colors.outline))
			}
		}
		highlightedID = 0
	})

	var lastX, lastY int

	c.BindItem("item", event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		lastX = ev.X
		lastY = ev.Y
	})

	c.BindItem("item", event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask == 0 {
			return
		}
		dx := float64(ev.X - lastX)
		dy := float64(ev.Y - lastY)
		c.Move("current", dx, dy)
		lastX = ev.X
		lastY = ev.Y
	})

	app.Run()
}

// findImage locates an image in the demos/images/ directory.
func findImage(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	demosRoot := filepath.Dir(filepath.Dir(file))
	path := filepath.Join(demosRoot, "images", name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

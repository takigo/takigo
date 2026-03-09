// Demo: Canvas widget with examples of the various canvas item types.
// Ported from Tk's items.tcl demo.
package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
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
		label.WrapLength("5i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window contains a canvas widget with examples of the various kinds of items supported by canvases.  The following operations are supported:\n  Left-Button drag:\tmoves item under pointer.\n  Middle-Button drag:\trepositions view.\n  Right-Button drag:\tstrokes out area.\n  Ctrl+f:\t\tprints items under area."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Outer grid frame for canvas + scrollbars.
	gf := frame.New(f, "frame")
	pack.Pack(gf, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Canvas with scroll region covering all sections including images, bitmaps, and windows.
	c := canvas.New(gf, "c",
		canvas.Width(screenunit.Px("15c")),
		canvas.Height(screenunit.Px("10c")),
		canvas.ScrollRegion(0, 0, screenunit.Px("30c"), screenunit.Px("24c")),
		canvas.ReliefOpt(option.ReliefSunken),
		canvas.BorderWidthOpt(2),
	)

	scrollCmd := func(viewFunc func(n int, pages bool), moveFunc func(f float64)) func(args ...any) {
		return func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						moveFunc(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					viewFunc(n, unit == "pages")
				}
			}
		}
	}

	yscroll := ttk.NewScrollbar(gf, "vscroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(scrollCmd(c.YViewScroll, c.YViewMoveTo)),
	)
	xscroll := ttk.NewScrollbar(gf, "hscroll",
		ttk.ScrollbarOrientOpt(ttk.Horizontal),
		ttk.ScrollbarCommandOpt(scrollCmd(c.XViewScroll, c.XViewMoveTo)),
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

	// Track original colors per item ID for hover restore.
	type itemColors struct {
		fill    string
		outline string
		isText  bool
	}
	origColors := map[string]itemColors{}

	record := func(id int64, fill, outline string) {
		origColors[fmt.Sprintf("%d", id)] = itemColors{fill: fill, outline: outline}
	}
	recordText := func(id int64, textColor string) {
		origColors[fmt.Sprintf("%d", id)] = itemColors{fill: textColor, isText: true}
	}

	// --- Section 1: Rectangles ---
	c.CreateText(110, 15, canvas.TextOpt("Rectangles"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreateRectangle(20, 30, 100, 90,
		canvas.FillColor("#4a86c8"), canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "#4a86c8", "black")

	record(c.CreateRectangle(120, 30, 200, 90,
		canvas.FillColor("#e8a835"), canvas.OutlineColor("black"), canvas.OutlineWidth(1),
		canvas.Tags("item")), "#e8a835", "black")

	record(c.CreateRectangle(20, 100, 200, 130,
		canvas.FillColor("#6bb86b"), canvas.OutlineColor("darkgreen"), canvas.OutlineWidth(3),
		canvas.Tags("item")), "#6bb86b", "darkgreen")

	// --- Section 2: Ovals ---
	c.CreateText(330, 15, canvas.TextOpt("Ovals"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreateOval(240, 30, 340, 90,
		canvas.FillColor("#c85a5a"), canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "#c85a5a", "black")

	record(c.CreateOval(360, 30, 420, 130,
		canvas.FillColor("#9b59b6"), canvas.OutlineColor("black"), canvas.OutlineWidth(1),
		canvas.Tags("item")), "#9b59b6", "black")

	// --- Section 3: Lines ---
	c.CreateText(550, 15, canvas.TextOpt("Lines"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreateLine([]float64{460, 30, 640, 30},
		canvas.OutlineColor("red"), canvas.OutlineWidth(3),
		canvas.Tags("item")), "", "red")

	record(c.CreateLine([]float64{460, 50, 550, 90, 640, 50},
		canvas.OutlineColor("blue"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "", "blue")

	record(c.CreateLine([]float64{460, 110, 550, 70, 640, 110},
		canvas.OutlineColor("darkgreen"), canvas.OutlineWidth(2), canvas.Smooth(true),
		canvas.Tags("item")), "", "darkgreen")

	record(c.CreateLine([]float64{460, 130, 640, 130},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.Arrow(canvas.ArrowBoth),
		canvas.Tags("item")), "", "black")

	// --- Section 4: Polygons ---
	c.CreateText(110, 155, canvas.TextOpt("Polygons"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreatePolygon([]float64{60, 170, 20, 260, 100, 260},
		canvas.FillColor("#e74c3c"), canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "#e74c3c", "black")

	cx, cy, r := 160.0, 220.0, 40.0
	penta := make([]float64, 10)
	for i := range 5 {
		angle := math.Pi/2 + float64(i)*2*math.Pi/5
		penta[i*2] = cx + r*math.Cos(angle)
		penta[i*2+1] = cy - r*math.Sin(angle)
	}
	record(c.CreatePolygon(penta,
		canvas.FillColor("#3498db"), canvas.OutlineColor("navy"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "#3498db", "navy")

	// --- Section 5: Arcs ---
	c.CreateText(330, 155, canvas.TextOpt("Arcs"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreateArc(240, 170, 340, 270,
		canvas.FillColor("#f39c12"), canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.StartAngle(0), canvas.Extent(120), canvas.ArcStyleOpt(canvas.ArcStylePieslice),
		canvas.Tags("item")), "#f39c12", "black")

	record(c.CreateArc(350, 170, 430, 270,
		canvas.OutlineColor("#c0392b"), canvas.OutlineWidth(3),
		canvas.StartAngle(30), canvas.Extent(270), canvas.ArcStyleOpt(canvas.ArcStyleArc),
		canvas.Tags("item")), "", "#c0392b")

	// --- Section 6: Text ---
	c.CreateText(550, 155, canvas.TextOpt("Text Items"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	recordText(c.CreateText(550, 190, canvas.TextOpt("Default font"),
		canvas.TextColor("black"), canvas.AnchorOpt(option.AnchorCenter),
		canvas.Tags("item")), "black")

	recordText(c.CreateText(550, 220, canvas.TextOpt("Bold text"),
		canvas.FontOpt("Sans Bold 14"), canvas.TextColor("navy"),
		canvas.AnchorOpt(option.AnchorCenter),
		canvas.Tags("item")), "navy")

	recordText(c.CreateText(550, 250, canvas.TextOpt("Italic text"),
		canvas.FontOpt("Sans Italic 12"), canvas.TextColor("darkred"),
		canvas.AnchorOpt(option.AnchorCenter),
		canvas.Tags("item")), "darkred")

	// --- Section 7: Dashed Lines ---
	c.CreateText(330, 290, canvas.TextOpt("Dashed Lines"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreateLine([]float64{20, 310, 300, 310},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2), canvas.Dash(6, 4),
		canvas.Tags("item")), "", "black")

	record(c.CreateLine([]float64{20, 330, 300, 330},
		canvas.OutlineColor("blue"), canvas.OutlineWidth(2), canvas.Dash(12, 4, 4, 4),
		canvas.Tags("item")), "", "blue")

	record(c.CreateRectangle(340, 300, 640, 380,
		canvas.OutlineColor("gray50"), canvas.OutlineWidth(2), canvas.Dash(8, 4),
		canvas.Tags("item")), "", "gray50")

	record(c.CreateOval(380, 305, 600, 375,
		canvas.FillColor("#eaf2f8"), canvas.OutlineColor("#2980b9"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "#eaf2f8", "#2980b9")

	// --- Section 8: Images ---
	c.CreateText(330, 430, canvas.TextOpt("Images"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	if imgPath := findImage("ouster.png"); imgPath != "" {
		if photo, err := tkimage.NewPhotoFromFile("items_ouster", imgPath); err == nil {
			app.ImageRegistry().Register(photo)
			c.CreateImage(110, 470, canvas.ImageOpt(photo),
				canvas.AnchorOpt(option.AnchorCenter),
				canvas.Tags("item"))
		}
	}
	if imgPath := findImage("plowed_field.png"); imgPath != "" {
		if photo, err := tkimage.NewPhotoFromFile("items_field", imgPath); err == nil {
			app.ImageRegistry().Register(photo)
			c.CreateImage(350, 470, canvas.ImageOpt(photo),
				canvas.AnchorOpt(option.AnchorCenter),
				canvas.Tags("item"))
		}
	}

	// --- Section 9: Bitmaps (XBM) ---
	c.CreateText(330, 560, canvas.TextOpt("Bitmaps (XBM)"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	bitmapNames := []string{"gray25.xbm", "flagdown.xbm", "flagup.xbm", "letters.xbm"}
	bx9 := 60.0
	for _, name := range bitmapNames {
		if bPath := findImage(name); bPath != "" {
			if xbm, err := canvas.ParseXBMFile(bPath); err == nil {
				c.CreateBitmap(bx9, 615, xbm,
					canvas.AnchorOpt(option.AnchorCenter),
					canvas.BitmapForeground(0, 0, 0),
					canvas.Tags("item"))
				bx9 += 160
			}
		}
	}

	// --- Section 10: Embedded Windows ---
	c.CreateText(330, 700, canvas.TextOpt("Embedded Windows"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	// A label widget embedded in the canvas.
	lbl := label.New(c, "win_label",
		label.Text("Canvas Label"),
		label.Relief(option.ReliefGroove),
		label.BorderWidth(2),
	)
	lbl.Win.ReqWidth = 120
	lbl.Win.ReqHeight = 30
	c.CreateWindow(80, 720, lbl.Win)

	// A button widget embedded in the canvas.
	btn := button.New(c, "win_button",
		button.Text("Click Me"),
		button.Command(func() {}),
	)
	btn.Win.ReqWidth = 100
	btn.Win.ReqHeight = 30
	c.CreateWindow(240, 720, btn.Win)

	// A frame with content embedded in the canvas.
	frm := frame.New(c, "win_frame",
		frame.Relief(option.ReliefSunken),
		frame.BorderWidth(2),
	)
	frm.Win.ReqWidth = 120
	frm.Win.ReqHeight = 30
	c.CreateWindow(380, 720, frm.Win)

	// --- Event bindings ---

	var highlightedID string

	c.BindItem("item", event.EnterMask, func(ev *event.Event) {
		highlightedID = ""
		ids := c.FindWithTag("current")
		if len(ids) == 0 {
			return
		}
		idStr := fmt.Sprintf("%d", ids[0])
		colors, ok := origColors[idStr]
		if !ok {
			return
		}
		highlightedID = idStr
		if colors.isText {
			c.ItemConfigure("current", canvas.TextColor("SteelBlue2"))
		} else if colors.fill != "" {
			c.ItemConfigure("current", canvas.FillColor("SteelBlue2"))
		} else if colors.outline != "" {
			c.ItemConfigure("current", canvas.OutlineColor("SteelBlue2"))
		}
	})

	c.BindItem("item", event.LeaveMask, func(ev *event.Event) {
		if highlightedID == "" {
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
		highlightedID = ""
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

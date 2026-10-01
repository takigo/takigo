// Demo: Text widget with embedded windows.
// Ported from Tk's twind.tcl demo.
package main

import (
	"fmt"
	"github.com/msorc/takigo/font"
	"os"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/panedwindow"
	"github.com/msorc/takigo/widget/text"
	"github.com/msorc/takigo/widget/toplevel"
)

var peerCount int

func makePeer(app *takigo.App, doc *text.Document) {
	peerCount++
	n := peerCount
	w := toplevel.New(app, fmt.Sprintf("peer%d", n))
	w.WmInfo.SetTitle(fmt.Sprintf("Text Peer #%d", n))

	pf := frame.New(w, "f",
		frame.BorderWidth(1),
		frame.Relief(option.ReliefSunken),
	)

	pt := text.NewPeer(doc, pf, "text",
		text.Width(70), text.Height(20),
		text.WrapModeOpt(text.WrapWord),
		text.BorderWidthOpt(0),
		text.HighlightThickness(0),
	)
	pack.Pack(pt, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	ps := ttk.NewScrollbar(w, "scroll")
	pt.YScrollCmd = func(first, last float64) { ps.Set(first, last) }
	ps.Command = func(args ...any) {
		if len(args) >= 2 {
			action, _ := args[0].(string)
			number, _ := args[1].(float64)
			switch action {
			case "moveto":
				pt.YViewMoveTo(number)
			case "scroll":
				unit := "units"
				if len(args) >= 3 {
					if u, ok := args[2].(string); ok {
						unit = u
					}
				}
				pt.YViewScroll(int(number), unit == "pages")
			}
		}
	}
	pack.Pack(ps, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(pf, pack.Expand(true), pack.FillOpt(pack.FillBoth))
	w.Show()
}

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

func main() {
	app, err := takigo.NewApp(takigo.Title("Text Demonstration - Embedded Windows and Other Features"),
		takigo.Geometry("+300+300"),
		takigo.IconName("Embedded Windows"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Panedwindow for potential split view (matches $w.pane in Tcl).
	pw := panedwindow.New(f, "pane",
		panedwindow.OrientOpt(panedwindow.Horizontal),
	)

	// Frame with sunken border to hold the main text widget (matches $w.f).
	tf := frame.New(pw, "tf",
		frame.HighlightThickness(1),
		frame.BorderWidth(1),
		frame.Relief(option.ReliefSunken),
	)

	tw := text.New(tf, "text",
		text.FontOpt(font.TkDefaultFont), // -font $font (mainFont)
		text.Width(70), text.Height(35),
		text.WrapModeOpt(text.WrapWord),
		text.BorderWidthOpt(0),
		text.HighlightThickness(0),
	)
	pack.Pack(tw, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	scroll := ttk.NewScrollbar(f, "scroll")
	tw.YScrollCmd = func(first, last float64) { scroll.Set(first, last) }
	scroll.Command = func(args ...any) {
		if len(args) >= 2 {
			action, _ := args[0].(string)
			number, _ := args[1].(float64)
			switch action {
			case "moveto":
				tw.YViewMoveTo(number)
			case "scroll":
				unit := "units"
				if len(args) >= 3 {
					if u, ok := args[2].(string); ok {
						unit = u
					}
				}
				tw.YViewScroll(int(number), unit == "pages")
			}
		}
	}
	pack.Pack(scroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))

	// Horizontal scrollbar (added/removed by Turn On/Off buttons).
	var hscroll *ttk.Scrollbar

	pw.Add(tf.Window(), 0)
	pw.SetStretch(tf.Window(), panedwindow.StretchAlways)
	pack.Pack(pw, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// Peer text widget for Split Windows (nil when not shown).
	var peerTw *text.TextWidget

	// Configure tags matching twind.tcl.
	tw.TagConfigure("center",
		text.TagJustify(option.JustifyCenter),
		text.TagSpacing1(screenunit.Mm(5)),
		text.TagSpacing3(screenunit.Mm(5)),
	)
	tw.TagConfigure("buttons",
		text.TagLMargin1(screenunit.Cm(1)),
		text.TagLMargin2(screenunit.Cm(1)),
		text.TagRMargin(screenunit.Cm(1)),
		text.TagSpacing1(screenunit.Mm(3)),
		text.TagSpacing2(0),
		text.TagSpacing3(0),
	)

	// Turn On / Turn Off buttons.
	onBtn := button.New(tw, "on",
		button.Text("Turn On"),
		button.Command(func() {
			if hscroll == nil {
				hscroll = ttk.NewScrollbar(f, "hscroll",
					ttk.ScrollbarOrientOpt(ttk.Horizontal),
				)
				tw.XScrollCmd = func(first, last float64) { hscroll.Set(first, last) }
				hscroll.Command = func(args ...any) {
					if len(args) >= 2 {
						action, _ := args[0].(string)
						number, _ := args[1].(float64)
						switch action {
						case "moveto":
							tw.XViewMoveTo(number)
						case "scroll":
							unit := "units"
							if len(args) >= 3 {
								if u, ok := args[2].(string); ok {
									unit = u
								}
							}
							tw.XViewScroll(int(number), unit == "pages")
						}
					}
				}
				pack.Pack(hscroll, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))
			}
			tw.SetWrapMode(text.WrapNone)
		}),
	)
	offBtn := button.New(tw, "off",
		button.Text("Turn Off"),
		button.Command(func() {
			if hscroll != nil {
				app.Server().UnmapWindow(hscroll.Win.PlatformID)
				hscroll = nil
				tw.XScrollCmd = nil
			}
			tw.SetWrapMode(text.WrapWord)
		}),
	)

	tw.Insert("end", "A text widget can contain many different kinds of items, ")
	tw.Insert("end", "both active and passive.  It can lay these out in various ")
	tw.Insert("end", "ways, with wrapping, tabs, centering, etc.  In addition, ")
	tw.Insert("end", "when the contents are too big for the window, smooth ")
	tw.Insert("end", "scrolling in all directions is provided.\n\n")

	tw.Insert("end", "A text widget can contain other widgets embedded ")
	tw.Insert("end", "it.  These are called \"embedded windows\", ")
	tw.Insert("end", "and they can consist of arbitrary widgets.  ")
	tw.Insert("end", "For example, here are two embedded button ")
	tw.Insert("end", "widgets.  You can click on the first button to ")
	tw.WindowCreate(tw.EndIndex(), onBtn.Window())
	tw.Insert("end", " horizontal scrolling, which also turns off ")
	tw.Insert("end", "word wrapping.  Or, you can click on the second ")
	tw.Insert("end", "button to\n")
	tw.WindowCreate(tw.EndIndex(), offBtn.Window())
	tw.Insert("end", " horizontal scrolling and turn back on word wrapping.\n\n")

	// Canvas plot section.
	var plotCanvas *canvas.Canvas

	tw.Insert("end", "Or, here is another example.  If you ")

	clickBtn := button.New(tw, "click",
		button.Text("Click Here"),
		button.Command(func() {
			if plotCanvas != nil {
				return
			}
			// Create canvas as child of text widget.
			plotCanvas = canvas.New(tw, "plotcanvas",
				canvas.Width(450), canvas.Height(300),
				canvas.BorderWidthOpt(2),
				canvas.ReliefOpt(option.ReliefSunken),
			)

			// Draw axes.
			plotCanvas.CreateLine([]float64{100, 250, 400, 250}, canvas.OutlineWidth(2))
			plotCanvas.CreateLine([]float64{100, 250, 100, 50}, canvas.OutlineWidth(2))
			plotCanvas.CreateText(225, 20, canvas.TextOpt("A Simple Plot"), canvas.TextColor("brown"))

			// X axis tick marks and labels.
			for i := 0; i <= 10; i++ {
				x := float64(100 + i*30)
				plotCanvas.CreateLine([]float64{x, 250, x, 245}, canvas.OutlineWidth(2))
				plotCanvas.CreateText(x, 258, canvas.TextOpt(fmt.Sprintf("%d", i*10)), canvas.AnchorOpt(option.AnchorN))
			}
			// Y axis tick marks and labels.
			for i := 0; i <= 5; i++ {
				y := float64(250 - i*40)
				plotCanvas.CreateLine([]float64{100, y, 105, y}, canvas.OutlineWidth(2))
				plotCanvas.CreateText(96, y, canvas.TextOpt(fmt.Sprintf("%d.0", i*50)), canvas.AnchorOpt(option.AnchorE))
			}

			// Data points.
			type pt struct{ x, y int }
			points := []pt{{12, 56}, {20, 94}, {33, 98}, {32, 120}, {61, 180}, {75, 160}, {98, 223}}
			for _, p := range points {
				x := float64(100 + 3*p.x)
				y := float64(250 - 4*p.y/5)
				plotCanvas.CreateOval(x-6, y-6, x+6, y+6,
					canvas.OutlineColor("black"),
					canvas.FillColor("SkyBlue2"),
					canvas.Tags("point"),
				)
			}

			// Hover highlight on data points.
			plotCanvas.BindItem("point", event.EnterMask, func(ev *event.Event) {
				id := plotCanvas.CurrentItem()
				if id >= 0 {
					plotCanvas.ItemConfigure(fmt.Sprintf("%d", id),
						canvas.FillColor("red"))
					plotCanvas.Display()
				}
			})
			plotCanvas.BindItem("point", event.LeaveMask, func(ev *event.Event) {
				id := plotCanvas.CurrentItem()
				if id >= 0 {
					plotCanvas.ItemConfigure(fmt.Sprintf("%d", id),
						canvas.FillColor("SkyBlue2"))
					plotCanvas.Display()
				}
			})

			// Drag support.
			var dragID int64 = -1
			var dragLastX, dragLastY int
			plotCanvas.BindItem("point", event.ButtonPressMask, func(ev *event.Event) {
				if ev.Button == 1 {
					dragID = plotCanvas.CurrentItem()
					dragLastX = ev.X
					dragLastY = ev.Y
				}
			})
			plotCanvas.BindItem("point", event.ButtonReleaseMask, func(ev *event.Event) {
				dragID = -1
			})
			app.Dispatcher().Bind(plotCanvas.Win.PlatformID, event.MotionMask, func(ev *event.Event) {
				if dragID < 0 || ev.State&platform.Button1Mask == 0 {
					return
				}
				dx := ev.X - dragLastX
				dy := ev.Y - dragLastY
				plotCanvas.Move(fmt.Sprintf("%d", dragID), float64(dx), float64(dy))
				dragLastX = ev.X
				dragLastY = ev.Y
				plotCanvas.Display()
			})

			// Insert the canvas at the "plot" mark and tag it as "center".
			tw.Insert("plot", "\n")
			tw.WindowCreate("plot", plotCanvas.Win)
			tw.TagAdd("center", "plot", "plot+1c")
			tw.Insert("plot", "\n")
			tw.Display()
		}),
	)
	tw.WindowCreate(tw.EndIndex(), clickBtn.Window())
	tw.Insert("end", " a canvas displaying an x-y plot will appear right here.")
	tw.MarkSet("plot", "insert")
	tw.MarkGravity("plot", text.GravityLeft)
	tw.Insert("end", "  You can drag the data points around with the mouse, ")
	tw.Insert("end", "or you can click here to ")

	deleteBtn := button.New(tw, "delete",
		button.Text("Delete"),
		button.Command(func() {
			if plotCanvas != nil {
				tw.RemoveWindow(plotCanvas.Win)
				plotCanvas.Destroy()
				plotCanvas = nil
				tw.Display()
			}
		}),
	)
	tw.WindowCreate(tw.EndIndex(), deleteBtn.Window())
	tw.Insert("end", " the plot again.\n\n")

	// Make A Peer / Split Windows section.
	doc := tw.Doc()
	makePeerBtn := button.New(tw, "makepeer",
		button.Text("Make A Peer"),
		button.Command(func() {
			makePeer(app, doc)
		}),
	)
	splitBtn := button.New(tw, "split",
		button.Text("Split Windows"),
		button.Command(func() {
			if peerTw != nil {
				pw.Remove(peerTw.Win)
				peerTw.Destroy()
				peerTw = nil
			} else {
				// "$textW peer create $w.peer -yscrollcommand ...": defaults otherwise.
				peerTw = text.NewPeer(doc, f, "peertext",
					text.Width(35), text.Height(35),
					text.WrapModeOpt(text.WrapWord),
				)
				pw.Add(peerTw.Win, 0)
				pw.SetStretch(peerTw.Win, panedwindow.StretchAlways)
			}
		}),
	)

	tw.Insert("end", "You can also create multiple text widgets each of which ")
	tw.Insert("end", "display the same underlying text. Click this button to ")
	tw.WindowCreatePad(tw.EndIndex(), makePeerBtn.Window(), screenunit.Pt(3), 0)
	tw.Insert("end", " widget.  Notice how peer widgets can have different ")
	tw.Insert("end", "font settings, and by default contain all the images ")
	tw.Insert("end", "of the 'parent', but that the embedded windows, ")
	tw.Insert("end", "such as buttons may not appear in the peer.  To ensure ")
	tw.Insert("end", "that embedded windows appear in all peers you can set the ")
	tw.Insert("end", "'-create' option to a script or a string containing %W.  ")
	tw.Insert("end", "(The plot above and the 'Make A Peer' button are ")
	tw.Insert("end", "designed to show up in all peers.)  A good use of ")
	tw.Insert("end", "peers is for ")
	tw.WindowCreatePad(tw.EndIndex(), splitBtn.Window(), screenunit.Pt(3), 0)
	tw.Insert("end", " \n\n")

	tw.Insert("end", "Users of previous versions of Tk will also be interested ")
	tw.Insert("end", "to note that now cursor movement is now by visual line by ")
	tw.Insert("end", "default, and that all scrolling of this widget is by pixel.\n\n")

	tw.Insert("end", "You may also find it useful to put embedded windows in ")
	tw.Insert("end", "a text without any actual text.  In this case the ")
	tw.Insert("end", "text widget acts like a geometry manager.  For ")
	tw.Insert("end", "example, here is a collection of buttons laid out ")
	tw.Insert("end", "neatly into rows by the text widget.  These buttons ")
	tw.Insert("end", "can be used to change the background color of the ")
	tw.Insert("end", "text widget (\"Default\" restores the color to ")
	tw.Insert("end", "its default).  If you click on the button labeled ")
	tw.Insert("end", "\"Short\", it changes to a longer string so that ")
	tw.Insert("end", "you can see how the text widget automatically ")
	tw.Insert("end", "changes the layout.  Click on the button again ")
	tw.Insert("end", "to restore the short string.\n")
	tw.Insert("end", "\nNOTE: these buttons will not appear in peers!\n")

	// Default button.
	defaultBtnLine := tw.EndIndex()
	defaultBtn := button.New(tw, "default",
		button.Text("Default"),
		button.Command(func() {
			tw.Configure(text.Background("#ffffff"))
		}),
	)
	tw.WindowCreatePad(defaultBtnLine, defaultBtn.Window(), screenunit.Pt(3), 0)

	// Toggle button "Short" / "A much longer string".
	var toggleLong bool
	var toggleBtn *button.Button
	toggleBtnLine := tw.EndIndex()
	toggleBtn = button.New(tw, "toggle",
		button.Text("Short"),
		button.Command(func() {
			toggleLong = !toggleLong
			if toggleLong {
				toggleBtn.SetText("A much longer string")
			} else {
				toggleBtn.SetText("Short")
			}
			tw.Display()
		}),
	)
	tw.WindowCreatePad(toggleBtnLine, toggleBtn.Window(), screenunit.Pt(3), screenunit.Pt(1.5))

	// Color buttons.
	colors := []string{
		"AntiqueWhite3", "Bisque1", "Bisque2", "Bisque3", "Bisque4",
		"SlateBlue3", "RoyalBlue1", "SteelBlue2", "DeepSkyBlue3", "LightBlue1",
		"DarkSlateGray1", "Aquamarine2", "DarkSeaGreen2", "SeaGreen1",
		"Yellow1", "IndianRed1", "IndianRed2", "Tan1", "Tan4",
	}
	for i, c := range colors {
		name := fmt.Sprintf("color%d", i+1)
		colorName := c
		btnLine := tw.EndIndex()
		clrBtn := button.New(tw, name,
			button.Text(colorName),
			button.Command(func() {
				tw.Configure(text.Background(colorName))
			}),
		)
		tw.WindowCreatePad(btnLine, clrBtn.Window(), screenunit.Pt(3), screenunit.Pt(1.5))
	}

	// Tag the buttons section.
	tw.TagAdd("buttons", defaultBtnLine, "end")

	// Border/highlight/pad buttons.
	tw.Insert("end", "\nYou can also change the usual border width and ")
	tw.Insert("end", "highlightthickness and padding.\n")

	normalBorder := tw.BorderWidth
	normalHighlight := tw.HighlightWidth
	normalPad := tw.PadX

	bigBBtn := button.New(tw, "bigB",
		button.Text("Big borders"),
		button.Command(func() {
			tw.Configure(text.BorderWidthOpt(12))
		}),
	)
	tw.WindowCreate(tw.EndIndex(), bigBBtn.Window())

	smallBBtn := button.New(tw, "smallB",
		button.Text("Small borders"),
		button.Command(func() {
			tw.Configure(text.BorderWidthOpt(normalBorder))
		}),
	)
	tw.WindowCreate(tw.EndIndex(), smallBBtn.Window())

	bigHBtn := button.New(tw, "bigH",
		button.Text("Big highlight"),
		button.Command(func() {
			tw.Configure(text.HighlightThickness(12))
		}),
	)
	tw.WindowCreate(tw.EndIndex(), bigHBtn.Window())

	smallHBtn := button.New(tw, "smallH",
		button.Text("Small highlight"),
		button.Command(func() {
			tw.Configure(text.HighlightThickness(normalHighlight))
		}),
	)
	tw.WindowCreate(tw.EndIndex(), smallHBtn.Window())

	bigPBtn := button.New(tw, "bigP",
		button.Text("Big pad"),
		button.Command(func() {
			tw.SetPadX(screenunit.Pt(12).Pixels())
			tw.SetPadY(screenunit.Pt(12).Pixels())
		}),
	)
	tw.WindowCreate(tw.EndIndex(), bigPBtn.Window())

	smallPBtn := button.New(tw, "smallP",
		button.Text("Small pad"),
		button.Command(func() {
			tw.SetPadX(normalPad)
			tw.SetPadY(normalPad)
		}),
	)
	tw.WindowCreate(tw.EndIndex(), smallPBtn.Window())

	// Image at end (matches $t image create end -image img2 in Tcl).
	tw.Insert("end", "\n\nFinally, images fit comfortably in text widgets too:")
	if imgPath := findImage("ouster.png"); imgPath != "" {
		if photo, err := tkimage.NewPhotoFromFile("twind_ouster", imgPath); err == nil {
			app.ImageRegistry().Register(photo)
			tw.ImageCreate(tw.EndIndex(), photo)
		}
	}

	app.Run()
}

// Demo: This demonstration showcases the tk print commands.
// Ported from Tk's print.tcl demo.
package main

import (
	"encoding/base64"
	"fmt"
	tkimage "github.com/takigo/takigo/image"
	"github.com/takigo/takigo/screenunit"
	goimage "image"
	"image/draw"
	"image/gif"
	"math"
	"os"
	"strings"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/dialog"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Printing Demonstration"),
		takigo.Geometry("+300+300"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "l",
		label.Text("This demonstration showcases\nthe tk print command. Clicking the buttons below\nprints the data from the canvas and text widgets\nusing platform-native dialogs."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	// Content area: canvas left, text right. Created before the button row
	// so the canvas is in scope when the print button's closure runs.
	m := frame.New(f, "m")
	pack.Pack(m, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	// Canvas with shapes.
	c := canvas.New(m, "c", canvas.Background("white"))
	pack.Pack(c, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	pt := screenunit.Distance.Float
	c.CreateRectangle(pt(screenunit.Pt(15)), pt(screenunit.Pt(15)), pt(screenunit.Pt(165)), pt(screenunit.Pt(60)),
		canvas.FillColor("blue"), canvas.OutlineColor("black"))
	c.CreateOval(pt(screenunit.Pt(15)), pt(screenunit.Pt(75)), pt(screenunit.Pt(165)), pt(screenunit.Pt(120)), canvas.FillColor("green"))
	logo, err := gif.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(logoData)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "print: %v\n", err)
		os.Exit(1)
	}
	// logo2 is logo zoomed by the integer scalingPct/100.
	zoom := max(1, screenunit.ScalingPct()/100)
	logo2 := tkimage.NewPhotoFromPhoto(tkimage.NewPhoto("logo", toRGBA(logo)), "logo2", tkimage.Zoom(float64(zoom)))
	imgID := c.CreateImage(pt(screenunit.Pt(90)), pt(screenunit.Pt(135)), canvas.ImageOpt(logo2), canvas.AnchorOpt(option.AnchorN))
	_, _, _, y2 := c.BBox(fmt.Sprint(imgID))
	y2 += int(math.Round(15 * screenunit.DPI() / 72)) // "15 pt to pixels" via [tk scaling]
	c.CreateText(pt(screenunit.Pt(15)), float64(y2), canvas.AnchorOpt(option.AnchorNW),
		canvas.FontOpt("Helvetica 12"), canvas.TextColor("black"),
		canvas.TextOpt("A short demo of simple canvas elements."))

	tw := text.New(m, "t", text.WrapModeOpt(text.WrapWord))
	pack.Pack(tw, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))
	tw.Insert("end", "\nTcl, or Tool Command Language, is an open-source multi-purpose C library which includes a powerful dynamic scripting language. Together they provide ideal cross-platform development environment for any programming project. It has served for decades as an essential system component in organizations ranging from NASA to Cisco Systems, is a must-know language in the fields of EDA, and powers companies such as FlightAware and F5 Networks.\n\nTcl is fit for both the smallest and largest programming tasks, obviating the need to decide whether it is overkill for a given job or whether a system written in Tcl will scale up as needed. Wherever a shell script might be used Tcl is a better choice, and entire web ecosystems and mission-critical control and testing systems have also been written in Tcl. Tcl excels in all these roles due to the minimal syntax of the language, the unique programming paradigm exposed at the script level, and the careful engineering that has gone into the design of the Tcl internals.\n")

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Button frame at the bottom.
	btnFrame := frame.New(f, "f")
	pack.Pack(btnFrame, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	printCanvasBtn := button.New(btnFrame, "c",
		button.Text("Print Canvas"),
		button.Command(func() {
			path := "/tmp/takigo-canvas.ps"
			_, err := c.Postscript(canvas.PSFile(path))
			if err != nil {
				dialog.ShowMessage(app,
					dialog.MsgTitle("Print"),
					dialog.MsgMessage("Failed: "+err.Error()),
					dialog.MsgType(dialog.MsgError),
				)
				return
			}
			dialog.ShowMessage(app,
				dialog.MsgTitle("Print"),
				dialog.MsgMessage("Saved to "+path),
				dialog.MsgType(dialog.MsgInfo),
			)
		}),
	)
	pack.Pack(printCanvasBtn, pack.SideOpt(pack.Left), pack.Anchor(option.AnchorW),
		pack.PadX(screenunit.Pt(3)))

	printTextBtn := button.New(btnFrame, "t",
		button.Text("Print Text"),
		button.Command(func() {
			// Text-widget printing is out of scope for the first port;
			// tk's full print.tcl driver layer builds on top of canvas
			// postscript + a platform print spooler. For now we just show
			// a "not yet" notice.
			dialog.ShowMessage(app,
				dialog.MsgTitle("Print"),
				dialog.MsgMessage("Text-widget printing is not yet implemented in takigo."),
				dialog.MsgType(dialog.MsgInfo),
			)
		}),
	)
	pack.Pack(printTextBtn, pack.SideOpt(pack.Right), pack.Anchor(option.AnchorE),
		pack.PadX(screenunit.Pt(3)))

	app.Run()
}

func toRGBA(img goimage.Image) *goimage.RGBA {
	b := img.Bounds()
	out := goimage.NewRGBA(goimage.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Bounds(), img, b.Min, draw.Src)
	return out
}

const logoData = "R0lGODlhMABLAPUAAP//////zP//mf//AP/MzP/Mmf/MAP+Zmf+ZZv+ZAMz//8zM/8zMzMyZzMyZmcyZZsyZAMxmZsxmM8xmAMwzM8wzAJnMzJmZzJmZmZlmmZlmZplmM5kzZpkzM5kzAGaZzGZmzGZmmWYzZmYzMzNmzDNmmTMzmTMzZgAzmQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACH+BSAtZGwtACH5BAEKAAIALAAAAAAwAEsAAAb+QIFwSCwahY9HRMI8Op/JJVNSqVqv2OvjyRU8slbIJGwYg60S5ZR6jRi/4ITBOhkYIOd8dltEnAdmFQMJeoVXCEd/VnKGjRVOZ3NVgHlsjpBxVRCEYBIEAAARl4lgZmVgEQAKFx8Mo0ZnpqgAFyi2JqKGmGebWRIAILbCIo27cYFWASTCtievRXqSVwQfzLYeeYESxlnSVRIW1igjWHJmjBXbpKXeFQTizlh1eJNVHbYf0LGc39XW2PIoVZE0whasWPSqFBBHrkKEA3QG0DFTEMXBUsjCWesg4oMFAGwgtKsiwqA+jGiCiRPGAM6pLCVLGKHQ6EGJlc0IuDxzAgX+CCOW9DjAaUsEyAoT+GHpeSRoHgxEUWgAUEUpFhMWgTbKEPUBAU15TBZxekYD0RMEqCDLIpYIWTAcmGEd9rWQBxQyjeQqdK/ZTWEO3mK5l+9No75SrcHhm9WwnlzNoA5zdM+JHz0HCPQdUauZowoFnSw+c2CBvw6dUXT4LMKE6EIHUqMexgCiIREknOwl7Q+FhNQoLuzOc6Kw3kIIVOLqjYKBYCwinmgo9CBEswfMAziK7mRDoQhcUZxwoBKFibq3n3jXI0GyCPLC0DrS8GR1oaEoRBRYVhT99/qG4DcCA/yNU4Ajbjhhnx4P2DJggR3YZog6RyyYxwM9PSgMBaP+sQdgIRL0JAKBwnTooRMAFWLdiPyJ8JwvTnyQoh5midCASh149ZkTIFAmHnzOZOBfIU6U4Mhd4zF34DNEoDAhARGY50BvJkioyxFOGkKAShGkFsJwejiR5Xf8aZAaBp89coQJjuDXAQOApekEm45ANaAtIbyYxREf0OlICCK841uaahZBQjyfjXCACYjuaASjhFagRKSFNtloHg+hYWIxRohnBQWCSSAhBVZ+hkgRnlbxwJIVgIqGlaU6wkeTxHxjm6gVLImrFbHWVEQ1taZjWxJX7KqqnqgUEUxDwtqajrOaRkqhEDcxWwECbEjxTYe9gojqOJQ6JO231ob72bSqAjh4RgfsjiDCCfDCK8K8I9TL7r33nvGtCO7CO1dUAONk3LcBFxzwwEMwZ/DC4iAsRIE+CWNCbzeV8FfEtoDwVwnlacxMkcKQYIE/F5TQ2QcedUZCagyc3NsFGrXVZMipWVBCzKv4Q0JvCviDsjAwf4ylxBeX0KcwGs81ccgqGS3MBxc3RjDDVAvdBRcfeFy1MFd3bcQHJEQdlddkP5E1Cf9yXfbaV2d9RBAAOw=="

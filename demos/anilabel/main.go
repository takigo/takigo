// Demo: Animated scrolling labels.
// Ported from Tk's anilabel.tcl demo.
package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	goimage "image"
	"image/draw"
	"image/gif"
	"os"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
)

// scrollLabel holds the state for one animated scrolling label.
type scrollLabel struct {
	label  *label.Label
	runes  []rune
	offset int
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Animated Label Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("anilabel"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(4)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Four animated labels are displayed below; each of the labels on the left is animated by making the text message inside it appear to scroll, and the label on the right is animated by animating the image that it displays."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Left labelframe: scrolling texts.
	leftFrame := labelframe.New(f, "left",
		labelframe.Text("Scrolling Texts"),
	)
	pack.Pack(leftFrame, pack.SideOpt(pack.Left), pack.PadX(screenunit.Pt(7.5)), pack.PadY(screenunit.Pt(7.5)),
		pack.Expand(true))

	// Right labelframe: GIF placeholder.
	rightFrame := labelframe.New(f, "right",
		labelframe.Text("GIF Image"),
	)
	pack.Pack(rightFrame, pack.SideOpt(pack.Left), pack.PadX(screenunit.Pt(7.5)), pack.PadY(screenunit.Pt(7.5)),
		pack.Expand(true))

	// Three scrolling labels with different messages and speeds,
	// matching the Tk original's l1 (slow, ridge), l2 (fast, groove), l3 (flat, long text).
	type labelSpec struct {
		name   string
		text   string
		relief option.Relief
		millis int
		width  int // -width in chars; 0 means omit
	}
	specs := []labelSpec{
		{"l1", "* Slow Animation *", option.ReliefRidge, 300, 0},
		{"l2", "* Fast Animation *", option.ReliefGroove, 80, 0},
		{"l3", "This is a longer scrolling text in a widget that will not show the whole message at once. ", option.ReliefFlat, 150, 18},
	}

	for _, spec := range specs {
		opts := []label.LabelOption{
			label.Text(spec.text),
			label.FontOpt(font.TkFixedFont),
			label.BorderWidth(screenunit.Pt(3).Pixels()),
			label.Relief(spec.relief),
		}
		if spec.width > 0 {
			opts = append(opts, label.Width(spec.width))
		}
		l := label.New(leftFrame, spec.name, opts...)
		pack.Pack(l, pack.SideOpt(pack.Top), pack.Expand(true),
			pack.PadX(screenunit.Pt(7.5)), pack.PadY(screenunit.Pt(7.5)), pack.Anchor(option.AnchorW))

		sl := &scrollLabel{
			label:  l,
			runes:  []rune(spec.text),
			offset: 0,
		}

		// Start animation for this label.
		interval := time.Duration(spec.millis) * time.Millisecond
		var animate func()
		animate = func() {
			sl.offset = (sl.offset + 1) % len(sl.runes)
			visible := make([]rune, len(sl.runes))
			for i := range sl.runes {
				visible[i] = sl.runes[(sl.offset+i)%len(sl.runes)]
			}
			sl.label.Configure(label.Text(string(visible)))
			app.After(interval, animate)
		}
		app.After(interval, animate)
	}

	// animateLabelImage: each GIF frame is copied over the previous ones,
	// as "$image2 copy $image" does after "$image configure -format".
	anim, err := gif.DecodeAll(base64.NewDecoder(base64.StdEncoding,
		bytes.NewReader([]byte(tclPoweredData))))
	if err != nil {
		fmt.Fprintf(os.Stderr, "anilabel: %v\n", err)
		os.Exit(1)
	}
	canvas := goimage.NewRGBA(goimage.Rect(0, 0, anim.Config.Width, anim.Config.Height))
	frameIdx := 0
	draw.Draw(canvas, anim.Image[0].Bounds(), anim.Image[0], anim.Image[0].Bounds().Min, draw.Over)
	photo := tkimage.NewPhoto("tclPowered", canvas)
	gifLabel := label.New(rightFrame, "l",
		label.BorderWidth(0),
		label.ImageOpt(photo),
	)
	pack.Pack(gifLabel, pack.SideOpt(pack.Top), pack.Expand(true),
		pack.PadX(screenunit.Pt(7.5)), pack.PadY(screenunit.Pt(7.5)))
	var nextFrame func()
	nextFrame = func() {
		frameIdx = (frameIdx + 1) % len(anim.Image)
		fr := anim.Image[frameIdx]
		draw.Draw(canvas, fr.Bounds(), fr, fr.Bounds().Min, draw.Over)
		photo.Invalidate()
		gifLabel.Display()
		app.After(100*time.Millisecond, nextFrame)
	}
	app.After(100*time.Millisecond, nextFrame)

	app.Run()
}

const tclPoweredData = "R0lGODlhKgBAAPQAAP//////zP//AP/MzP/Mmf/MAP+Zmf+ZZv+ZAMz//8zMzMyZmcyZZsxmZsxmAMwzAJnMzJmZzJmZmZlmmZlmZplmM5kzM2aZzGZmzGZmmWZmZmYzZmYzMzNmzDMzZgAzmSH+IE1hZGUgd2l0aCBHSU1QIGJ5IExARGVtYWlsbHkuY29tACH5BAVkAAEALAAAAAAqAEAAAAX+YCCOZEkyTKM2jOm66yPPdF03bx7YcuHIDkGBR7SZeIyhTID4FZ+4Es8nQyCe2EeUNJ0peY2s9mi7PhAMngEAMGRbUpvzSxskLh1J+Hkg134OdDIDEB+GHxtYMEQMTjMGEYeGFoomezaCDZGSHFmLXTQKkh8eNQVpZ2afmDQGHaOYSoEyhhcklzVmMpuHnaZmDqiGJbg0qFqvh6UNAwB7VA+OwydEjgujkgrPNhbTI8dFvNgEYcHcHx0lB1kX2IYeA2G6NN0YfkXJ2BsAMuAzHB9cZMk3qoEbRzUACsRCUBK5JxsC3iMiKd8GN088SIyT0RAFSROyeEg38caDiB/+JEgqxsODrZJ1BkT0oHKSmI0ceQxo94HDpg0qsuDkUmRAMgu8OgwQ+uIJgUMVeGXA+IQkzEeHGvD8cIGlDXsLiRjQ+EHroQhea7xY8IQBSgYYDi1IS+OFBCgaDMGVS3fGi5BPJpBaENdQ0EomKGD56IHwO39EXiSCYsgxor5+Xfgq0qByYUpiXmwuoredB2aYH4gWWda0B7SeNENpEJHC1ghi+pS4AJpIAwWvKPBi+8YEht5EriEqpFfMlhEdkBNpx0HUhwypx5T4IB1MBg/Ws2snwV3MSQOkzI8fUd48Aw3dOZto71x85hHtHijYv18Gf/3GqCdDCXHNoICBobSoIqBqJLyCoH8JPrLgdh88CKCFD0CGmAiGYPgffwceZh6FC2ohIIklnkhehTNY4CIHHGzgwYw01ujBBhvAqKOLLq5AAk9kuSPkkKO40NB+h1gnypJIIvkBf09aN5QIRz5p5ZJXJpmlIVhOGQA2TmIJZZhKKmmll2BqyWSXWUrZpQtpatlmk1c2KaWRHeTZEJF8SqLDn/hhsOeQgBbqAh6DGqronxeARUIIACH5BAUeAAAALAUALgAFAAUAAAUM4CeKz/OV5YmqaRkCACH5BAUeAAEALAUALgAKAAUAAAUUICCKz/OdJVCaa7p+7aOWcDvTZwgAIfkEBR4AAQAsCwAuAAkABQAABRPgA4zP95zAeZqoWqqpyqLkZ38hACH5BAUKAAEALAcALgANAA4AAAU7ICA+jwiUJEqeKau+r+vGaTmac63v/GP9HM7GQyx+jsgkkoRUHJ3Qx0cK/VQVTKtWwbVKn9suNuncWkMAIfkEBQoAAAAsBwA3AAcABQAABRGgIHzk842j+Yjlt5KuO8JmCAAh+QQFCgAAACwLADcABwAFAAAFEeAnfN9TjqP5oOWziq05lmUIACH5BAUKAAAALA8ANwAHAAUAAAUPoPCJTymS3yiQj4qOcPmEACH5BAUKAAAALBMANwAHAAUAAAURoCB+z/MJX2o+I2miKimiawgAIfkEBQoAAAAsFwA3AAcABQAABRGgIHzfY47jQ4qk+aHl+pZmCAAh+QQFCgAAACwbADcABwAFAAAFEaAgfs/zCV9qPiNJouo7ll8IACH5BAUKAAAALB8ANwADAAUAAAUIoCB8o0iWZggAOw=="

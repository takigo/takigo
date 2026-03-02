// Demo: Built-in bitmap patterns displayed as generated images.
// Ported from Tk's bitmap.tcl demo.
package main

import (
	goimage "image"
	"image/color"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("Bitmap Demonstration", 600, 350,
		"Tk defines several built-in bitmap patterns. Below are\nrepresentations of the 10 standard bitmaps as images.")
	root, app := d.Root, d.App

	fg := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	bg := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Gray pattern: fill density% of pixels with fg.
	makeGray := func(name string, density int) *tkimage.Photo {
		const sz = 32
		img := goimage.NewRGBA(goimage.Rect(0, 0, sz, sz))
		step := 100 / density
		for y := range sz {
			for x := range sz {
				if (x+y*3)%step == 0 {
					img.SetRGBA(x, y, fg)
				} else {
					img.SetRGBA(x, y, bg)
				}
			}
		}
		return tkimage.NewPhoto(name, img)
	}

	// Simple icon: filled rect with letter.
	makeIcon := func(name string, letter byte, c color.RGBA) *tkimage.Photo {
		const sz = 32
		img := goimage.NewRGBA(goimage.Rect(0, 0, sz, sz))
		for y := range sz {
			for x := range sz {
				if x == 0 || x == sz-1 || y == 0 || y == sz-1 {
					img.SetRGBA(x, y, fg)
				} else {
					img.SetRGBA(x, y, c)
				}
			}
		}
		// Draw letter in center (simple 5x7 block).
		cx, cy := sz/2-2, sz/2-3
		for dy := 0; dy < 7; dy++ {
			for dx := 0; dx < 5; dx++ {
				if dx == 0 || dx == 4 || dy == 0 || dy == 3 || dy == 6 {
					img.SetRGBA(cx+dx, cy+dy, fg)
				}
			}
		}
		_ = letter
		return tkimage.NewPhoto(name, img)
	}

	type bitmapDef struct {
		name  string
		photo *tkimage.Photo
	}

	bitmaps := []bitmapDef{
		{"error", makeIcon("bm_error", 'E', color.RGBA{R: 220, G: 60, B: 60, A: 255})},
		{"gray12", makeGray("bm_gray12", 12)},
		{"gray25", makeGray("bm_gray25", 25)},
		{"gray50", makeGray("bm_gray50", 50)},
		{"gray75", makeGray("bm_gray75", 75)},
		{"hourglass", makeIcon("bm_hourglass", 'H', color.RGBA{R: 200, G: 180, B: 100, A: 255})},
		{"info", makeIcon("bm_info", 'I', color.RGBA{R: 60, G: 120, B: 220, A: 255})},
		{"questhead", makeIcon("bm_questhead", 'Q', color.RGBA{R: 100, G: 180, B: 100, A: 255})},
		{"question", makeIcon("bm_question", '?', color.RGBA{R: 60, G: 160, B: 220, A: 255})},
		{"warning", makeIcon("bm_warning", 'W', color.RGBA{R: 220, G: 180, B: 40, A: 255})},
	}

	for _, b := range bitmaps {
		app.ImageRegistry().Register(b.photo)
	}

	// Row 1: first 5 bitmaps.
	row1 := frame.New(root, "row1", app)
	pack.Pack(row1, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(10))

	for _, b := range bitmaps[:5] {
		col := frame.New(row1.Window(), "col_"+b.name, app)
		pack.Pack(col, pack.SideOpt(pack.Left), pack.PadX(10))
		il := label.New(col.Window(), "img_"+b.name, app,
			label.ImageOpt(b.photo),
			label.BorderWidth(2), label.Relief(option.ReliefGroove),
			label.PadX(4), label.PadY(4),
		)
		pack.Pack(il, pack.SideOpt(pack.Top))
		nl := label.New(col.Window(), "name_"+b.name, app, label.Text(b.name))
		pack.Pack(nl, pack.SideOpt(pack.Top), pack.PadY(2))
		_ = il
		_ = nl
	}

	// Row 2: last 5 bitmaps.
	row2 := frame.New(root, "row2", app)
	pack.Pack(row2, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(10))

	for _, b := range bitmaps[5:] {
		col := frame.New(row2.Window(), "col_"+b.name, app)
		pack.Pack(col, pack.SideOpt(pack.Left), pack.PadX(10))
		il := label.New(col.Window(), "img_"+b.name, app,
			label.ImageOpt(b.photo),
			label.BorderWidth(2), label.Relief(option.ReliefGroove),
			label.PadX(4), label.PadY(4),
		)
		pack.Pack(il, pack.SideOpt(pack.Top))
		nl := label.New(col.Window(), "name_"+b.name, app, label.Text(b.name))
		pack.Pack(nl, pack.SideOpt(pack.Top), pack.PadY(2))
		_ = il
		_ = nl
	}

	d.Run()
}

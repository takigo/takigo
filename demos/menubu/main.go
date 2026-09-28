// Demo: Menus and cascaded menus using menubuttons.
// Ported from Tk's menubu.tcl demo.
package main

import (
	"fmt"
	goimage "image"
	gocolor "image/color"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/menubutton"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Menu Button Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("menubutton"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Body frame — expands to fill.
	body := frame.New(f, "body")
	pack.Pack(body, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// Bottom buttons.
	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Helper to create a menubutton with a 2-item menu.
	makeMB := func(parent widget.Caregiver, name, text string, dir menubutton.Direction) *menubutton.Menubutton {
		mb := menubutton.New(parent, name,
			menubutton.Text(text),
			menubutton.UnderlineOpt(0),
			menubutton.DirectionOpt(dir),
			menubutton.Relief(option.ReliefRaised),
		)
		m := menu.New(mb, "m")
		m.AddCommand(text+" menu: first item", func() {
			fmt.Printf("You have selected the first item from the %s menu.\n", text)
		})
		m.AddCommand(text+" menu: second item", func() {
			fmt.Printf("You have selected the second item from the %s menu.\n", text)
		})
		mb.Menu = m
		return mb
	}

	// Compass grid layout:
	//   row 0, col 1: below (sticky n)
	//   row 1, col 0: right (sticky w)
	//   row 1, col 1: center frame
	//   row 1, col 2: left  (sticky e)
	//   row 2, col 1: above (sticky s)

	mbBelow := makeMB(body, "below", "Below", menubutton.Below)
	grid.Grid(mbBelow, grid.Row(0), grid.Column(1), grid.Sticky(grid.StickN))

	mbRight := makeMB(body, "right", "Right", menubutton.Right)
	center := frame.New(body, "center")
	mbLeft := makeMB(body, "left", "Left", menubutton.Left)
	grid.Grid(mbRight, grid.Row(1), grid.Column(0), grid.Sticky(grid.StickW))
	grid.Grid(center, grid.Row(1), grid.Column(1), grid.Sticky(grid.NSEW))
	grid.Grid(mbLeft, grid.Row(1), grid.Column(2), grid.Sticky(grid.StickE))

	mbAbove := makeMB(body, "above", "Above", menubutton.Above)
	grid.Grid(mbAbove, grid.Row(2), grid.Column(1), grid.Sticky(grid.StickS))

	// Center label.
	centerLabel := label.New(center, "label",
		label.WrapLength("225p"),
		label.FontOpt("Helvetica 14"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This is a demonstration of menubuttons. The \"Below\" menubutton pops its menu below the button; the \"Right\" button pops to the right, etc. There are two option menus directly below this text; one is just a standard menu and the other is a 16-color palette."),
	)
	pack.Pack(centerLabel, pack.SideOpt(pack.Top), pack.PadX("18p"), pack.PadY("18p"))

	// Option menu buttons frame.
	buttons := frame.New(center, "buttons")
	pack.Pack(buttons, pack.PadX("18p"), pack.PadY("18p"))

	// Simple option menu: one / two / three.
	optVar := widget.NewVariable("one")
	optMB := menubutton.New(buttons, "options",
		menubutton.Text(optVar.Get()),
		menubutton.OptionMenuOpt(true),
	)
	optMenu := menu.New(optMB, "m")
	for _, val := range []string{"one", "two", "three"} {
		v := val
		optMenu.AddCommand(v, func() {
			optVar.Set(v)
			optMB.SetText(v)
		})
	}
	optMB.Menu = optMenu
	pack.Pack(optMB, pack.SideOpt(pack.Left), pack.PadX("18p"), pack.PadY("18p"))

	// makeColorSwatch creates normal and selected 16x16 color swatch images.
	makeColorSwatch := func(colorName, topBorder, bottomBorder string) (*tkimage.Photo, *tkimage.Photo) {
		resolve := func(name string) gocolor.RGBA {
			col, err := app.ColorCache().Get(name)
			if err != nil {
				return gocolor.RGBA{128, 128, 128, 255}
			}
			return gocolor.RGBA{
				uint8(col.Red >> 8),
				uint8(col.Green >> 8),
				uint8(col.Blue >> 8),
				255,
			}
		}
		fill := resolve(colorName)
		top := resolve(topBorder)
		bot := resolve(bottomBorder)
		black := gocolor.RGBA{0, 0, 0, 255}

		// Normal swatch: top/left border = top color, bottom/right = bottomBorder, interior = fill.
		normRGBA := goimage.NewRGBA(goimage.Rect(0, 0, 16, 16))
		for x := range 16 {
			normRGBA.SetRGBA(x, 0, top)
		}
		for y := 1; y < 16; y++ {
			normRGBA.SetRGBA(0, y, top)
		}
		for x := range 16 {
			normRGBA.SetRGBA(x, 15, bot)
		}
		for y := 1; y < 15; y++ {
			normRGBA.SetRGBA(15, y, bot)
		}
		for y := 1; y < 15; y++ {
			for x := 1; x < 15; x++ {
				normRGBA.SetRGBA(x, y, fill)
			}
		}
		normPhoto := tkimage.NewPhoto(colorName+"_norm", normRGBA)
		app.ImageRegistry().Register(normPhoto)

		// Selected swatch: 2px black border, interior = fill.
		selRGBA := goimage.NewRGBA(goimage.Rect(0, 0, 16, 16))
		for x := range 16 {
			selRGBA.SetRGBA(x, 0, black)
			selRGBA.SetRGBA(x, 1, black)
			selRGBA.SetRGBA(x, 14, black)
			selRGBA.SetRGBA(x, 15, black)
		}
		for y := 2; y < 14; y++ {
			selRGBA.SetRGBA(0, y, black)
			selRGBA.SetRGBA(1, y, black)
			selRGBA.SetRGBA(14, y, black)
			selRGBA.SetRGBA(15, y, black)
		}
		for y := 2; y < 14; y++ {
			for x := 2; x < 14; x++ {
				selRGBA.SetRGBA(x, y, fill)
			}
		}
		selPhoto := tkimage.NewPhoto(colorName+"_sel", selRGBA)
		app.ImageRegistry().Register(selPhoto)

		return normPhoto, selPhoto
	}

	// 16-color palette option menu with image swatches in a 4-column grid.
	type paletteEntry struct {
		name     string
		colBreak bool
	}
	paletteEntries := []paletteEntry{
		{"Black", true}, {"red4", false}, {"DarkGreen", false}, {"NavyBlue", false},
		{"gray75", true}, {"Red", false}, {"Green", false}, {"Blue", false},
		{"gray50", true}, {"Yellow", false}, {"Cyan", false}, {"Magenta", false},
		{"White", true}, {"Brown", false}, {"DarkSeaGreen", false}, {"DarkViolet", false},
	}
	paletteVar := widget.NewVariable("Black")
	paletteMB := menubutton.New(buttons, "colors",
		menubutton.Text(paletteVar.Get()),
		menubutton.OptionMenuOpt(true),
	)
	paletteMenu := menu.New(paletteMB, "m", menu.TearOffOpt(true))
	for _, pe := range paletteEntries {
		norm, sel := makeColorSwatch(pe.name, "gray50", "gray75")
		c := pe.name
		paletteMenu.AddImageSwatch(norm, sel, pe.colBreak, func() {
			paletteVar.Set(c)
			paletteMB.SetText(c)
		})
	}
	paletteMB.Menu = paletteMenu
	pack.Pack(paletteMB, pack.SideOpt(pack.Left), pack.PadX("18p"), pack.PadY("18p"))

	grid.ColumnConfigure(body, 1, grid.Weight(1))
	grid.RowConfigure(body, 1, grid.Weight(1))

	_ = mbBelow
	_ = mbLeft
	_ = mbRight
	_ = mbAbove
	_ = center
	_ = centerLabel
	_ = optVar
	_ = paletteVar
	app.Run()
}

// Demo: Simple collection of widgets to select and view images in a Tk label.
// Ported from Tk's image2.tcl demo.
package main

import (
	"fmt"
	goimage "image"
	_ "image/jpeg" // register JPEG decoder
	"os"
	"path/filepath"
	"sort"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/listbox"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Image Demonstration #2"),
		takigo.Geometry("+300+300"),
		takigo.IconName("Image2"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This demonstration allows you to view images using a Tk \"photo\" image.  First type a directory name in the listbox, then type Return to load the directory into the listbox.  Then double-click on a file name in the listbox to see that image."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Icon label at bottom (shows icon for selected image).
	iconLabel := label.New(f, "label",
		label.Text("Icon for Selected Image"),
		label.Relief(option.ReliefGroove),
	)
	pack.Pack(iconLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Find demos/images directory relative to working directory.
	dirName := findImagesDir()

	// Middle frame.
	mid := frame.New(f, "mid")
	pack.Pack(mid, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// --- "Directory:" labelframe ---
	dirLF := labelframe.New(mid, "dir", labelframe.Text("Directory:"))
	dirEntry := entry.New(dirLF, "e", entry.Width(30))
	dirEntry.SetText(dirName)

	var lb *listbox.Listbox

	// loadDir reloads the directory listbox from the directory named in the entry.
	loadDir := func() {
		dir := dirEntry.GetText()
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		// Clear listbox.
		n := lb.ItemCount()
		if n > 0 {
			lb.Delete(0, n-1)
		}
		// List all files (matching Tcl's glob -type f).
		var files []string
		for _, de := range entries {
			if !de.IsDir() {
				files = append(files, de.Name())
			}
		}
		sort.Strings(files)
		for _, f := range files {
			lb.Insert(lb.ItemCount(), f)
		}
	}

	// selectAndLoadDir pops up a directory chooser dialog and reloads the list.
	selectAndLoadDir := func() {
		dir, ok := dialog.ChooseDirectory(f,
			dialog.DirTitle("Select a directory"),
			dialog.DirInitialDir(dirEntry.GetText()),
			dialog.DirMustExist(true),
		)
		if ok {
			dirEntry.SetText(dir)
			loadDir()
		}
	}

	selectDirBtn := button.New(dirLF, "b",
		button.Text("Select Dir."),
		button.PadX("2m"), button.PadY(0),
		button.Command(selectAndLoadDir),
	)
	pack.Pack(dirEntry, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.PadX("2m"), pack.PadY("2m"), pack.Expand(true))
	pack.Pack(selectDirBtn, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY),
		pack.PadXPair(0, "2m"), pack.PadY("2m"))

	// --- "File:" labelframe ---
	fileLF := labelframe.New(mid, "f", labelframe.Text("File:"),
		labelframe.PadX("2m"), labelframe.PadY("2m"))
	lb = listbox.New(fileLF, "list",
		listbox.Width(20),
		listbox.Height(10),
	)
	yscroll := ttk.NewScrollbar(fileLF, "scroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						lb.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					lb.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	lb.YScrollCmd = func(first, last float64) { yscroll.Set(first, last) }
	pack.Pack(lb, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.Expand(true))
	pack.Pack(yscroll, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.Expand(true))

	// Register widgets with the bind engine so bindings work.
	eng := app.Bind()
	eng.RegisterWindow(dirEntry.Window(), "Entry")
	eng.RegisterWindow(lb.Window(), "Listbox")

	// Bind Return on entry to load directory.
	eng.Bind(dirEntry.Window().PathName, "<Return>", func(_ *bind.EventData) bool {
		loadDir()
		return false
	})

	// The list starts with these names; the directory is only read on
	// Return or "Select Dir.".
	for _, name := range []string{"earth.gif", "earthris.gif", "teapot.ppm", "Tcl.svg"} {
		lb.Insert(lb.ItemCount(), name)
	}

	// --- "Image:" labelframe ---
	imageLF := labelframe.New(mid, "image", labelframe.Text("Image:"))
	// "image create photo image2a": an empty photo until a file is loaded.
	imgLabel := label.New(imageLF, "image",
		label.ImageOpt(tkimage.NewPhoto("image2a", goimage.NewRGBA(goimage.Rect(0, 0, 0, 0)))))
	pack.Pack(imgLabel, pack.PadX("2m"), pack.PadY("2m"))

	// Double-click on listbox loads the image.
	var currentPhotoName string
	eng.Bind(lb.Window().PathName, "<Double-Button-1>", func(_ *bind.EventData) bool {
		app.DoWhenIdle(func() {
			sel := lb.Selection()
			if len(sel) == 0 {
				return
			}
			items := lb.GetItems()
			if sel[0] >= len(items) {
				return
			}
			filename := items[sel[0]]
			dir := dirEntry.GetText()
			path := filepath.Join(dir, filename)

			// Free previous photos.
			if currentPhotoName != "" {
				app.ImageRegistry().Unregister(currentPhotoName)
			}

			photoName := fmt.Sprintf("img2a_%s", filename)
			newPhoto, err := tkimage.NewPhotoFromFile(photoName, path)
			if err != nil {
				// Mark the file as not loadable with red background.
				lb.ItemConfigure(sel[0], "", "#c00000")
				return
			}
			app.ImageRegistry().Register(newPhoto)
			currentPhotoName = newPhoto.Name()

			// Show the full image in the "Image:" labelframe.
			imgLabel.Configure(label.Text(""), label.ImageOpt(newPhoto))

			// "tk fileicon $filename 48".
			iconLabel.Configure(label.CompoundOpt(widget.CompoundTop),
				label.ImageOpt(demohelper.FileIcon(path, 48)))
		})
		return false
	})

	// Grid: dir spans 2 cols row 0; f and image on row 1 (matches Tcl's grid layout).
	grid.Grid(dirLF, grid.Row(0), grid.Column(0), grid.ColumnSpan(2),
		grid.Sticky(grid.EW), grid.PadX("1m"), grid.PadY("1m"))
	grid.Grid(fileLF, grid.Row(1), grid.Column(0),
		grid.Sticky(grid.StickN|grid.StickW), grid.PadX("1m"), grid.PadY("1m"))
	grid.Grid(imageLF, grid.Row(1), grid.Column(1),
		grid.Sticky(grid.StickN|grid.StickW), grid.PadX("1m"), grid.PadY("1m"))
	grid.ColumnConfigure(mid, 1, grid.Weight(1))

	app.Run()
}

// findImagesDir returns the path to the demos/images directory.
func findImagesDir() string {
	candidates := []string{
		"demos/images",
		"../images",
		"images",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			if abs, err := filepath.Abs(c); err == nil {
				return abs
			}
		}
	}
	cwd, _ := os.Getwd()
	return cwd
}

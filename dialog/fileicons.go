package dialog

import (
	goimage "image"
	"image/color"

	tkimage "github.com/takigo/takigo/image"
)

// fileIcons are the 16x16 pictograms of the file dialog, drawn in code so
// the package ships no image assets (Tk embeds its folder/file/updir GIFs
// in tkfbox.tcl the same way).
type fileIcons struct {
	folder, file, up, home, newFolder, drive *tkimage.Photo
}

func newFileIcons() *fileIcons {
	return &fileIcons{
		folder:    tkimage.NewPhoto("", drawIcon(paintFolder)),
		file:      tkimage.NewPhoto("", drawIcon(paintFile)),
		up:        tkimage.NewPhoto("", drawIcon(paintUp)),
		home:      tkimage.NewPhoto("", drawIcon(paintHome)),
		newFolder: tkimage.NewPhoto("", drawIcon(paintNewFolder)),
		drive:     tkimage.NewPhoto("", drawIcon(paintDrive)),
	}
}

func (ic *fileIcons) destroy() {
	for _, p := range []*tkimage.Photo{ic.folder, ic.file, ic.up, ic.home, ic.newFolder, ic.drive} {
		p.Destroy()
	}
}

const iconSize = 16

type raster struct{ img *goimage.RGBA }

func drawIcon(paint func(r raster)) *goimage.RGBA {
	r := raster{goimage.NewRGBA(goimage.Rect(0, 0, iconSize, iconSize))}
	paint(r)
	return r.img
}

func (r raster) rect(x0, y0, x1, y1 int, c color.RGBA) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			r.img.SetRGBA(x, y, c)
		}
	}
}

// poly fills a convex polygon given its vertices clockwise on screen
// (y grows downwards, so the edge cross product is positive inside).
func (r raster) poly(c color.RGBA, pts ...[2]float64) {
	n := len(pts)
	for y := range iconSize {
		for x := range iconSize {
			px, py := float64(x)+0.5, float64(y)+0.5
			inside := true
			for i := range n {
				ax, ay := pts[i][0], pts[i][1]
				bx, by := pts[(i+1)%n][0], pts[(i+1)%n][1]
				if (bx-ax)*(py-ay)-(by-ay)*(px-ax) < 0 {
					inside = false
					break
				}
			}
			if inside {
				r.img.SetRGBA(x, y, c)
			}
		}
	}
}

var (
	folderDark  = color.RGBA{0xc8, 0x8a, 0x1e, 0xff}
	folderLight = color.RGBA{0xf5, 0xc0, 0x4e, 0xff}
	folderTab   = color.RGBA{0xe0, 0xa4, 0x33, 0xff}
	paperEdge   = color.RGBA{0x8a, 0x8a, 0x8a, 0xff}
	paperFill   = color.RGBA{0xfc, 0xfc, 0xfc, 0xff}
	paperLine   = color.RGBA{0xb8, 0xb8, 0xb8, 0xff}
	inkDark     = color.RGBA{0x3c, 0x3c, 0x3c, 0xff}
	accentGreen = color.RGBA{0x2e, 0x9e, 0x44, 0xff}
	driveGrey   = color.RGBA{0x9a, 0x9a, 0x9a, 0xff}
	driveLight  = color.RGBA{0xd8, 0xd8, 0xd8, 0xff}
	homeRoof    = color.RGBA{0xb0, 0x3a, 0x2e, 0xff}
	homeWall    = color.RGBA{0xf0, 0xe6, 0xc8, 0xff}
)

func paintFolder(r raster) {
	r.rect(1, 3, 7, 5, folderTab)
	r.rect(1, 5, 15, 14, folderDark)
	r.rect(2, 7, 14, 13, folderLight)
}

func paintNewFolder(r raster) {
	paintFolder(r)
	r.rect(7, 8, 9, 13, accentGreen)
	r.rect(5, 10, 11, 12, accentGreen)
	r.rect(7, 9, 9, 12, paperFill)
	r.rect(6, 10, 10, 11, paperFill)
	r.rect(7, 8, 9, 13, accentGreen)
	r.rect(5, 10, 11, 12, accentGreen)
}

func paintFile(r raster) {
	r.rect(3, 1, 10, 15, paperEdge)
	r.rect(9, 5, 13, 15, paperEdge)
	r.rect(4, 2, 9, 14, paperFill)
	r.rect(9, 6, 12, 14, paperFill)
	r.poly(paperEdge, [2]float64{9, 1}, [2]float64{13, 5}, [2]float64{9, 5})
	r.poly(paperFill, [2]float64{9, 2}, [2]float64{12, 5}, [2]float64{9, 5})
	for _, y := range []int{7, 9, 11} {
		r.rect(5, y, 11, y+1, paperLine)
	}
}

func paintUp(r raster) {
	r.poly(inkDark, [2]float64{8, 1}, [2]float64{15, 8}, [2]float64{1, 8})
	r.rect(6, 8, 10, 15, inkDark)
}

func paintHome(r raster) {
	r.poly(homeRoof, [2]float64{8, 1}, [2]float64{15.5, 8}, [2]float64{0.5, 8})
	r.rect(3, 8, 13, 15, homeWall)
	r.rect(3, 8, 13, 9, homeRoof)
	r.rect(7, 10, 10, 15, inkDark)
}

func paintDrive(r raster) {
	r.rect(1, 3, 15, 13, driveGrey)
	r.rect(2, 4, 14, 8, driveLight)
	r.rect(2, 9, 14, 12, driveLight)
	r.rect(12, 10, 13, 11, accentGreen)
}

package canvas

import (
	"strconv"
	"testing"
	"unicode/utf8"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// benchServer is a DisplayServer whose drawing calls are no-ops; methods
// the canvas does not call are left to the nil embedded interface.
type benchServer struct{ platform.DisplayServer }

func (benchServer) MapWindow(platform.WindowID)                                            {}
func (benchServer) UnmapWindow(platform.WindowID)                                          {}
func (benchServer) MoveResizeWindow(platform.WindowID, int, int, uint, uint)               {}
func (benchServer) FillRectangle(platform.DrawableID, platform.GCID, int, int, uint, uint) {}
func (benchServer) DrawRectangle(platform.DrawableID, platform.GCID, int, int, uint, uint) {}
func (benchServer) DrawLine(platform.DrawableID, platform.GCID, int, int, int, int)        {}
func (benchServer) DrawLines(platform.DrawableID, platform.GCID, []platform.Point, int)    {}
func (benchServer) FillPolygon(platform.DrawableID, platform.GCID, []platform.Point, int, int) {
}
func (benchServer) FillArc(platform.DrawableID, platform.GCID, int, int, uint, uint, int, int) {}
func (benchServer) DrawArc(platform.DrawableID, platform.GCID, int, int, uint, uint, int, int) {}
func (benchServer) CopyArea(platform.DrawableID, platform.DrawableID, platform.GCID, int, int, uint, uint, int, int) {
}
func (benchServer) SetDashes(platform.GCID, int, []byte)                 {}
func (benchServer) SetForeground(platform.GCID, uint64)                  {}
func (benchServer) SetLineAttributes(platform.GCID, uint, int, int, int) {}
func (benchServer) SetFillStyle(platform.GCID, int)                      {}
func (benchServer) SetArcMode(platform.GCID, int)                        {}
func (benchServer) SetStipple(platform.GCID, platform.PixmapID)          {}
func (benchServer) SetTSOrigin(platform.GCID, int, int)                  {}
func (benchServer) CreatePixmap(platform.DrawableID, uint, uint, uint) platform.PixmapID {
	return 1
}
func (benchServer) FreePixmap(platform.PixmapID) {}
func (benchServer) CreateBitmapFromData(platform.DrawableID, []byte, uint, uint) platform.PixmapID {
	return 2
}
func (benchServer) Flush() {}

// benchApp is the AppContext of a display-free canvas: idle callbacks are
// dropped, since the benchmarks paint explicitly.
type appContext = widget.AppContext

type benchApp struct{ appContext }

func (benchApp) DoWhenIdle(func())              {}
func (benchApp) Server() platform.DisplayServer { return benchDisplay }

// fixedFont is a display-free font: 7px per rune, ascent 10, descent 3.
type fixedFont struct{}

func (fixedFont) Attrs() font.Attributes     { return font.Attributes{} }
func (fixedFont) Metrics() font.Metrics      { return font.Metrics{Ascent: 10, Descent: 3, MaxWidth: 7} }
func (fixedFont) MeasureString(s string) int { return 7 * utf8.RuneCountInString(s) }
func (fixedFont) Close()                     {}
func (fixedFont) DrawString(platform.DrawableID, int, int, string, uint64, uint16, uint16, uint16) {
}

const benchW, benchH = 800, 600

// benchDisplay is the one interface value the benchmarks draw through, so
// no per-call conversion of benchServer is measured.
var benchDisplay platform.DisplayServer = benchServer{}

func newBenchCanvas() *Canvas {
	c := newCanvas()
	c.Win = &window.Window{Width: benchW, Height: benchH, Depth: 24,
		Display: &window.Display{Server: benchDisplay}}
	c.Win.Flags |= window.FlagMapped
	c.App = benchApp{}
	c.inset = 1
	c.xOrigin, c.yOrigin = c.inset, c.inset
	return c
}

func benchText(c *Canvas, x, y float64, text string) *TextItem {
	t := &TextItem{x: x, y: y, anchor: option.AnchorCenter, font: fixedFont{}, text: text}
	t.color = &color.ColorRef{}
	t.ItemBase.canvas = c
	t.updateBBox()
	return t
}

// fillPixel sets an item's fill without a colour cache.
func fillPixel(p uint64) ItemOption {
	return func(_ *Canvas, item Item) error {
		switch it := item.(type) {
		case *RectOvalItem:
			it.fill = &color.ColorRef{Pixel: p}
		case *PolygonItem:
			it.fill = &color.ColorRef{Pixel: p}
		}
		return nil
	}
}

// benchScene fills c with n items of mixed types spread over a 2000x2000
// area, so a 800x600 view sees roughly an eighth of them, each tagged
// "tag<i%10>". Every tenth item is a smooth polygon, every seventh a line
// with arrowheads, every fifth a text.
func benchScene(c *Canvas, n int) []ItemID {
	ids := make([]ItemID, 0, n)
	for i := range n {
		x := float64((i * 97) % 2000)
		y := float64((i * 61) % 2000)
		tag := Tags("tag" + strconv.Itoa(i%10))
		var id ItemID
		switch i % 5 {
		case 0:
			id = c.createItem(newRectOvalItem("rectangle", x, y, x+40, y+30, c),
				[]ItemOption{tag, fillPixel(0xff0000)})
		case 1:
			id = c.createItem(newRectOvalItem("oval", x, y, x+40, y+30, c), []ItemOption{tag})
		case 2:
			opts := []ItemOption{tag, WidthOpt(3)}
			if i%7 == 0 {
				opts = append(opts, Arrow(ArrowBoth))
			}
			id = c.createItem(newLineItem([]float64{x, y, x + 20, y + 40, x + 40, y, x + 60, y + 40}, c), opts)
		case 3:
			opts := []ItemOption{tag, fillPixel(0x00ff00)}
			if i%10 == 3 {
				opts = append(opts, Smooth(true))
			}
			id = c.createItem(newPolygonItem([]float64{x, y, x + 30, y - 10, x + 50, y + 20, x + 30, y + 50, x, y + 40}, c), opts)
		case 4:
			id = c.addItem(benchText(c, x, y, "item "+strconv.Itoa(i)))
			c.AddTag("tag"+strconv.Itoa(i%10), id)
		}
		ids = append(ids, id)
	}
	c.hasDamage, c.damageAll = false, false
	return ids
}

func BenchmarkPickCurrentItem(b *testing.B) {
	c := newBenchCanvas()
	benchScene(c, 1000)
	b.Run("miss", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			c.pickCurrentItem(-500, -500)
		}
	})
	b.Run("hit", func(b *testing.B) {
		// The first item's rectangle (0,0)-(40,30); it is at the bottom of
		// the display list, so the search visits every item above it.
		b.ReportAllocs()
		for range b.N {
			c.pickCurrentItem(20, 15)
		}
	})
}

func BenchmarkFindClosest(b *testing.B) {
	c := newBenchCanvas()
	benchScene(c, 1000)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		c.FindClosest(1000, 1000, 0, "")
	}
}

func BenchmarkFindOverlapping(b *testing.B) {
	c := newBenchCanvas()
	benchScene(c, 1000)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		c.Find("overlapping", 500, 500, 900, 900)
	}
}

func BenchmarkResolve(b *testing.B) {
	c := newBenchCanvas()
	ids := benchScene(c, 1000)
	id := ids[500]
	b.Run("id", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			c.resolve(id)
		}
	})
	b.Run("tag", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			c.resolve("tag3")
		}
	})
	b.Run("all", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			c.resolve("all")
		}
	})
}

func BenchmarkCreateDelete(b *testing.B) {
	c := newBenchCanvas()
	benchScene(c, 1000)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		id := c.CreateRectangle(10, 10, 50, 50, Tags("new"))
		c.Delete(id)
	}
}

func BenchmarkDeleteAll(b *testing.B) {
	// Deleting every item one by one, as a redraw-from-scratch demo does.
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		c := newBenchCanvas()
		ids := benchScene(c, 1000)
		b.StartTimer()
		for _, id := range ids {
			c.Delete(id)
		}
	}
}

func BenchmarkRaiseLower(b *testing.B) {
	c := newBenchCanvas()
	ids := benchScene(c, 1000)
	id := ids[500]
	b.Run("raise-id", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			c.Raise(id)
		}
	})
	b.Run("lower-id", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			c.Lower(id)
		}
	})
	b.Run("raise-tag", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			c.Raise("tag3")
		}
	})
}

func BenchmarkMove(b *testing.B) {
	c := newBenchCanvas()
	benchScene(c, 1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		d := float64(1 - 2*(i&1))
		c.Move("tag3", d, d)
	}
}

func BenchmarkSetItemCoords(b *testing.B) {
	c := newBenchCanvas()
	benchScene(c, 100)
	id := c.CreateLine(make([]float64, 400))
	coords := make([]float64, 400)
	for i := range coords {
		coords[i] = float64(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = c.SetItemCoords(id, coords)
	}
}

func BenchmarkPaint(b *testing.B) {
	c := newBenchCanvas()
	benchScene(c, 1000)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		c.paint(c.xOrigin, c.yOrigin, c.xOrigin+benchW-2, c.yOrigin+benchH-2)
	}
}

func BenchmarkItemDisplay(b *testing.B) {
	c := newBenchCanvas()
	d, drawable, gc := benchDisplay, platform.PixmapDrawable(1), platform.GCID(0)
	coords := make([]float64, 40)
	for i := range coords {
		coords[i] = float64(i * 13 % 97)
	}
	items := []struct {
		name string
		item Item
	}{
		{"rectangle", newRectOvalItem("rectangle", 10, 10, 200, 100, c)},
		{"line", newLineItem(coords, c)},
		{"line-arrows", func() Item {
			l := newLineItem(coords, c)
			l.arrow = ArrowBoth
			return l
		}()},
		{"line-smooth", func() Item {
			l := newLineItem(coords, c)
			l.smooth = true
			return l
		}()},
		{"polygon", newPolygonItem(coords, c)},
		{"polygon-smooth", func() Item {
			p := newPolygonItem(coords, c)
			p.smooth = true
			return p
		}()},
		{"text", benchText(c, 100, 100, "The quick brown fox\njumps over\tthe lazy dog")},
	}
	for _, it := range items {
		b.Run(it.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				it.item.Display(d, drawable, gc, -30, -30, benchW+60, benchH+60, -30, -30)
			}
		})
	}
}

func BenchmarkTextTranslate(b *testing.B) {
	c := newBenchCanvas()
	t := benchText(c, 100, 100, "The quick brown fox\njumps over\tthe lazy dog")
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		t.Translate(float64(1-2*(i&1)), 0)
	}
}

func BenchmarkPointDistance(b *testing.B) {
	c := newBenchCanvas()
	coords := make([]float64, 40)
	for i := range coords {
		coords[i] = float64(i * 13 % 97)
	}
	items := []struct {
		name string
		item Item
	}{
		{"line", newLineItem(coords, c)},
		{"polygon", newPolygonItem(coords, c)},
		{"oval", newRectOvalItem("oval", 0, 0, 100, 50, c)},
	}
	for _, it := range items {
		b.Run(it.name, func(b *testing.B) {
			for range b.N {
				it.item.PointDistance(150, 150)
			}
		})
	}
}

func BenchmarkSpline(b *testing.B) {
	coords := make([]float64, 40)
	for i := range coords {
		coords[i] = float64(i * 13 % 97)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		generateBezierSpline(coords, true, 12)
	}
}

func BenchmarkDispatchItemEvent(b *testing.B) {
	c := newBenchCanvas()
	ids := benchScene(c, 100)
	n := 0
	c.BindItem("tag3", event.ButtonPressMask, func(*event.Event) { n++ })
	c.BindItem(ids[3], event.ButtonPressMask, func(*event.Event) { n++ })
	c.currentItem = c.idMap[ids[3]]
	ev := &event.Event{Type: event.ButtonPressType}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		c.dispatchItemEvent(ev)
	}
	if n != 2*b.N {
		b.Fatalf("handlers ran %d times, want %d", n, 2*b.N)
	}
}

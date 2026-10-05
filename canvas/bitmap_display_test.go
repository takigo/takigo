package canvas_test

import (
	"testing"
	"time"

	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
)

// TestBitmapItemDrawsThroughMask checks a bitmap with a transparent
// background: set bits take the foreground and clear bits show the item
// underneath, as DisplayBitmap draws them.
func TestBitmapItemDrawsThroughMask(t *testing.T) {
	app := testutil.NewTestApp(t)
	c := canvas.New(app, "c", canvas.Width(60), canvas.Height(40),
		canvas.Background("white"), canvas.BorderWidthOpt(0), canvas.HighlightWidthOpt(0))
	pack.Pack(geometry.Group{c})

	// 8x2: first row all set, second row clear.
	xbm, err := canvas.ParseXBM("#define t_width 8\n#define t_height 2\nstatic char t_bits[] = { 0xff, 0x00 };")
	if err != nil {
		t.Fatal(err)
	}
	c.CreateRectangle(0, 0, 60, 40, canvas.FillColor("red"), canvas.OutlineWidth(0))
	c.CreateBitmap(10, 10, xbm, canvas.AnchorOpt(option.AnchorNW), canvas.BitmapForeground(0, 0, 255))

	var set, clear [3]byte
	app.After(300*time.Millisecond, func() {
		defer app.Quit()
		c.Display()
		app.UpdateIdleTasks()
		d := app.Server()
		d.Sync(false)
		px := d.GetImageRGBA(platform.WindowDrawable(c.Window().PlatformID), 12, 10, 1, 2)
		if px == nil {
			return
		}
		copy(set[:], px[0:3])
		copy(clear[:], px[4:7])
	})
	app.MainLoop()
	if set == ([3]byte{}) && clear == ([3]byte{}) {
		t.Skip("window contents cannot be read back")
	}
	if set != [3]byte{0, 0, 255} {
		t.Errorf("set bit drawn as %v, want blue", set)
	}
	if clear != [3]byte{255, 0, 0} {
		t.Errorf("clear bit drawn as %v, want the red item beneath", clear)
	}
}

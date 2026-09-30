package widget_test

import (
	"testing"

	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/text"
)

func TestConfigureRelaysOutParent(t *testing.T) {
	app := testutil.NewTestApp(t)
	f := frame.New(app, "f")
	pack.Pack(f)
	l := label.New(f, "l", label.Text("x"))
	b := button.New(f, "b", button.Text("x"))
	pack.Pack(l)
	pack.Pack(b)
	app.UpdateIdleTasks()
	labelW, buttonW, frameW := l.Win.ReqWidth, b.Win.ReqWidth, f.Win.ReqWidth

	l.Configure(label.Text("a considerably longer label text"))
	b.SetText("a considerably longer button text")
	if l.Win.ReqWidth <= labelW || b.Win.ReqWidth <= buttonW {
		t.Fatalf("requests did not grow: label %d -> %d, button %d -> %d", labelW, l.Win.ReqWidth, buttonW, b.Win.ReqWidth)
	}
	app.UpdateIdleTasks()
	if f.Win.ReqWidth <= frameW {
		t.Errorf("frame request %d did not follow its content (was %d)", f.Win.ReqWidth, frameW)
	}

	frameW = f.Win.ReqWidth
	f.Configure(frame.BorderWidth(10), frame.Background("red"))
	app.UpdateIdleTasks()
	if f.Win.ReqWidth != frameW+20 {
		t.Errorf("BorderWidth(10): frame request = %d, want %d", f.Win.ReqWidth, frameW+20)
	}
	if f.Win.BackgroundPixel != f.Background.Pixel || f.Background.Red != 0xffff {
		t.Errorf("Background(red): pixel %x, colour %+v", f.Win.BackgroundPixel, f.Background)
	}
}

func TestTextConfigureInsets(t *testing.T) {
	app := testutil.NewTestApp(t)
	tw := text.New(app, "t", text.Width(10), text.Height(2))
	pack.Pack(tw)
	w, h := tw.Win.ReqWidth, tw.Win.ReqHeight
	tw.Configure(text.BorderWidthOpt(tw.BorderWidth+5), text.PadXOpt(tw.PadX+3))
	if tw.Win.ReqWidth != w+16 || tw.Win.ReqHeight != h+10 {
		t.Errorf("request = %dx%d, want %dx%d", tw.Win.ReqWidth, tw.Win.ReqHeight, w+16, h+10)
	}
}

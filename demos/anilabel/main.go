// Demo: Animated scrolling labels.
// Ported from Tk's anilabel.tcl demo.
package main

import (
	"time"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
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
	app := demohelper.Setup("Animated Labels", 500, 300,
		"Four animated labels are displayed below; each of the labels on the left is animated by making the text message inside it appear to scroll, and the label on the right is animated by animating the image that it displays.")

	// Left labelframe: scrolling texts.
	leftFrame := labelframe.New(app, "left",
		labelframe.Text("Scrolling Texts"),
		labelframe.BorderWidth(2),
		labelframe.Relief(option.ReliefGroove),
	)
	pack.Pack(leftFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.PadX(10), pack.PadY(10), pack.Expand(true))

	// Right labelframe: GIF placeholder.
	rightFrame := labelframe.New(app, "right",
		labelframe.Text("GIF Image"),
		labelframe.BorderWidth(2),
		labelframe.Relief(option.ReliefGroove),
	)
	pack.Pack(rightFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.PadX(10), pack.PadY(10), pack.Expand(true))

	// Three scrolling labels with different messages and speeds,
	// matching the Tk original's l1 (slow, ridge), l2 (fast, groove), l3 (flat, long text).
	type labelSpec struct {
		name    string
		text    string
		relief  option.Relief
		millis  int
		fixedW  bool
	}
	specs := []labelSpec{
		{"l1", "* Slow Animation *", option.ReliefRidge, 300, false},
		{"l2", "* Fast Animation *", option.ReliefGroove, 80, false},
		{"l3", "This is a longer scrolling text in a widget that will not show the whole message at once. ", option.ReliefFlat, 150, true},
	}

	var scrollLabels []*scrollLabel
	for _, spec := range specs {
		opts := []label.LabelOption{
			label.Text(spec.text),
			label.Anchor(option.AnchorW),
			label.FontOpt("Courier 12"),
			label.BorderWidth(4),
			label.Relief(spec.relief),
		}
		if spec.fixedW {
			opts = append(opts, label.Width(180))
		}
		l := label.New(leftFrame, spec.name, opts...)
		pack.Pack(l, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadX(10), pack.PadY(10))

		sl := &scrollLabel{
			label:  l,
			runes:  []rune(spec.text),
			offset: 0,
		}
		scrollLabels = append(scrollLabels, sl)

		// Start animation for this label.
		interval := time.Duration(spec.millis) * time.Millisecond
		var animate func()
		animate = func() {
			sl.offset = (sl.offset + 1) % len(sl.runes)
			visible := make([]rune, len(sl.runes))
			for i := range sl.runes {
				visible[i] = sl.runes[(sl.offset+i)%len(sl.runes)]
			}
			sl.label.Text = string(visible)
			sl.label.Display()
			app.After(interval, animate)
		}
		app.After(interval, animate)
	}

	// Placeholder label in right frame for animated GIF.
	gifPlaceholder := label.New(rightFrame, "gif",
		label.Text("(Animated GIF\nnot supported)"),
		label.Anchor(option.AnchorCenter),
		label.PadX(20), label.PadY(40),
	)
	pack.Pack(gifPlaceholder, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.PadX(10), pack.PadY(10), pack.Expand(true))

	app.Run()
}

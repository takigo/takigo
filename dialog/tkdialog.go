package dialog

import (
	"fmt"

	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

// TkDialog holds the widgets of a tk_dialog-style window.
type TkDialog struct {
	Top, Bot *frame.Frame
	Msg      *label.Label
	Bitmap   *label.Label // nil when no bitmap was given
	Buttons  []*button.Button
}

// BuildTkDialog fills w with the contents of Tk's tk_dialog
// (tk/library/dialog.tcl, x11 branch): a raised top frame with the bitmap
// and the message (TkCaptionFont, wrapped at 3i) and a raised bottom frame
// with one classic button per label, def being the -default active one
// (-1 for none). onPress receives the index of the pressed button.
//
// Tk packs the message and bitmap with -in; takigo's pack has no -in, so
// they are created as children of the frames, which looks the same.
func BuildTkDialog(w widget.Caregiver, text, bitmap string, def int,
	buttons []string, onPress func(int)) *TkDialog {
	d := &TkDialog{}
	d.Bot = frame.New(w, "bot", frame.Relief(option.ReliefRaised), frame.BorderWidth(1))
	d.Top = frame.New(w, "top", frame.Relief(option.ReliefRaised), frame.BorderWidth(1))
	pack.Pack(d.Bot, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillBoth))
	pack.Pack(d.Top, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	grid.SetAnchor(d.Bot, option.AnchorCenter)

	d.Msg = label.New(d.Top, "msg",
		label.JustifyOpt(option.JustifyLeft),
		label.Text(text),
		label.WrapLength(screenunit.In(3)),
		label.FontOpt(font.TkCaptionFont),
	)
	pack.Pack(d.Msg, pack.SideOpt(pack.Right), pack.Expand(true), pack.FillOpt(pack.FillBoth),
		pack.PadX(screenunit.Mm(3)), pack.PadY(screenunit.Mm(3)))
	if bitmap != "" {
		d.Bitmap = label.New(d.Top, "bitmap", label.Bitmap(bitmap))
		pack.Pack(d.Bitmap, pack.SideOpt(pack.Left), pack.PadX(screenunit.Mm(3)), pack.PadY(screenunit.Mm(3)))
	}

	for i, text := range buttons {
		state := button.DefaultNormal
		if i == def {
			state = button.DefaultActive
		}
		b := button.New(d.Bot, fmt.Sprintf("button%d", i),
			button.Text(text),
			button.Default(state),
			button.Command(func() {
				if onPress != nil {
					onPress(i)
				}
			}),
		)
		grid.Grid(b, grid.Column(i), grid.Row(0), grid.Sticky(grid.EW),
			grid.PadX(screenunit.Pt(7.5)), grid.PadY(screenunit.Pt(3)))
		d.Buttons = append(d.Buttons, b)
	}
	return d
}

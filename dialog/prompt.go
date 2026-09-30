package dialog

import (
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget"
)

// AskString displays a modal dialog with a prompt and a text entry.
// Returns the entered text and true, or "" and false if cancelled.
func AskString(parent widget.Caregiver, title, prompt, initial string) (string, bool) {
	d := New(parent, title, 360, 120)

	body := ttk.NewFrame(d.Content, "body", ttk.FramePadding(ttk.UniformPadding(12)))
	pack.Pack(body, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	lab := ttk.NewLabel(body, "prompt", ttk.LabelText(prompt))
	pack.Pack(lab, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(4))

	ent := ttk.NewEntry(body, "entry", ttk.EntryWidth(36), ttk.EntryText(initial))
	pack.Pack(ent, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(4))
	_ = ent.Selection().Range("0", "end")

	addButtons(d, []dialogButton{
		{text: "OK", result: ResultOK, isDefault: true},
		{text: "Cancel", result: ResultCancel},
	})
	// The entry, not the OK button, gets the focus; Return still accepts.
	d.defaultButton = ent.Window()

	if d.Run() == ResultOK {
		return ent.Get(), true
	}
	return "", false
}

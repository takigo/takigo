package ttk

import (
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/panedwindow"
)

// Panedwindow is a TTK-themed panedwindow with a flat sash style.
// It wraps the classic PanedWindow but renders a thin flat divider
// instead of a 3D raised sash, matching Tcl's ttk::panedwindow appearance.
type Panedwindow struct {
	*panedwindow.PanedWindow
}

// PanedwindowOption is an alias for the classic PanedWindowOption.
type PanedwindowOption = panedwindow.PanedWindowOption

// NewPanedwindow creates a TTK-styled panedwindow with a flat sash.
func NewPanedwindow(parent widget.Caregiver, name string, opts ...PanedwindowOption) *Panedwindow {
	pw := panedwindow.New(parent, name, opts...)
	pw.Win.Class = "TPanedwindow"
	pw.FlatSash = true
	pw.Weighted = true
	pw.BorderWidth = 0 // ttk::panedwindow has no border
	pw.SashWidth = screenunit.Pt(3.75).Pixels()
	pw.GripSize = screenunit.Pt(15).Pixels()
	return &Panedwindow{PanedWindow: pw}
}

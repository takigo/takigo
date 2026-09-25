package ttk

// indicatorLayout ports the default theme's Checkbutton/Radiobutton layout:
//
//	padding { indicator -side left ; focus -side left -sticky w { label } }
//
// The indicator is -indicatorsize square (IndicatorElementSize, default 16)
// plus -indicatormargin; the focus element adds 1px around the label.
type indicatorLayout struct {
	pad, margin Padding
	size        int
}

const focusThickness = 1

func newIndicatorLayout(style *Style, state State) indicatorLayout {
	l := indicatorLayout{size: 16}
	if style != nil {
		l.pad = LookupPadding(style, "-padding", state, Padding{})
		l.margin = LookupPadding(style, "-indicatormargin", state, Padding{})
		l.size = LookupInt(style, "-indicatorsize", state, 16)
	}
	return l
}

func (l indicatorLayout) indicatorBox() (w, h int) {
	return l.size + l.margin.Width(), l.size + l.margin.Height()
}

// reqSize is the widget's requested size for a textW x textH label.
func (l indicatorLayout) reqSize(textW, textH int) (int, int) {
	indW, indH := l.indicatorBox()
	labelW, labelH := textW+2*focusThickness, textH+2*focusThickness
	return l.pad.Width() + indW + labelW, l.pad.Height() + max(indH, labelH)
}

// place returns the indicator's top-left corner and the label's top-left
// corner in a winW x winH window, both parcels centred vertically.
func (l indicatorLayout) place(winH, textH int) (indX, indY, textX, textY int) {
	indW, indH := l.indicatorBox()
	avail := winH - l.pad.Height()
	indX = l.pad.Left + l.margin.Left
	indY = l.pad.Top + (avail-indH)/2 + l.margin.Top
	labelH := textH + 2*focusThickness
	textX = l.pad.Left + indW + focusThickness
	textY = l.pad.Top + (avail-labelH)/2 + focusThickness
	return indX, indY, textX, textY
}

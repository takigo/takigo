package widget

// ScrollRequest is what a scrollbar asks of the widget it scrolls: Tk's
// "moveto fraction" and "scroll number units|pages" view commands.
type ScrollRequest struct {
	// MoveTo asks for the view to start at Fraction of the content.
	MoveTo   bool
	Fraction float64
	// Otherwise the view moves by Count units, or pages when Pages is set.
	Count int
	Pages bool
}

// ScrollTo returns the request Tk writes as "moveto fraction".
func ScrollTo(fraction float64) ScrollRequest {
	return ScrollRequest{MoveTo: true, Fraction: fraction}
}

// ScrollUnits returns the request Tk writes as "scroll n units".
func ScrollUnits(n int) ScrollRequest { return ScrollRequest{Count: n} }

// ScrollPages returns the request Tk writes as "scroll n pages".
func ScrollPages(n int) ScrollRequest { return ScrollRequest{Count: n, Pages: true} }

// YScrollable is a widget whose view scrolls vertically.
type YScrollable interface {
	YViewMoveTo(fraction float64)
	YViewScroll(count int, pages bool)
}

// XScrollable is a widget whose view scrolls horizontally.
type XScrollable interface {
	XViewMoveTo(fraction float64)
	XViewScroll(count int, pages bool)
}

// ScrollY returns a scrollbar command that scrolls w vertically:
//
//	sb := scrollbar.New(parent, "sb", scrollbar.CommandOpt(widget.ScrollY(list)))
func ScrollY(w YScrollable) func(ScrollRequest) {
	return func(r ScrollRequest) {
		if r.MoveTo {
			w.YViewMoveTo(r.Fraction)
		} else {
			w.YViewScroll(r.Count, r.Pages)
		}
	}
}

// ScrollX returns a scrollbar command that scrolls w horizontally.
func ScrollX(w XScrollable) func(ScrollRequest) {
	return func(r ScrollRequest) {
		if r.MoveTo {
			w.XViewMoveTo(r.Fraction)
		} else {
			w.XViewScroll(r.Count, r.Pages)
		}
	}
}

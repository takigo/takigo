package widget

// Configurable is a widget that Configure can reconfigure: any type
// embedding Base with a Display method.
type Configurable interface {
	Display()
	WidgetBase() *Base
}

// Configure applies opts to w and then does what Tk's WorldChanged procs do
// after any configure: rebuild the border, recompute the geometry, keep the
// window background in step, re-arrange the content when the margins
// changed, re-request the size from the geometry manager when it changed,
// and schedule a redraw. computeGeometry may be nil.
func Configure[W Configurable, O ~func(W)](w W, opts []O, computeGeometry func()) {
	b := w.WidgetBase()
	win := b.Win
	reqW, reqH := win.ReqWidth, win.ReqHeight
	insets := win.ContentInsets()
	for _, opt := range opts {
		opt(w)
	}
	b.UpdateBorder()
	if computeGeometry != nil {
		computeGeometry()
	}
	if b.Background != nil {
		win.SetBackgroundPixel(b.Background.Pixel)
	}
	if win.ContentInsets() != insets {
		win.NotifyConfigure()
	}
	if (win.ReqWidth != reqW || win.ReqHeight != reqH) && win.GeomManager != nil {
		win.GeomManager.RequestProc(win)
	}
	w.Display()
}

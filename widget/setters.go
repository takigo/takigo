package widget

import (
	"errors"
	"log/slog"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/font"
)

// WidgetBase returns the embedded Base; Configure reaches the common
// fields of any widget through it.
func (b *Base) WidgetBase() *Base { return b }

// OptionErrors collects the errors of options that could not be applied.
// Inside a Configure call they are gathered and returned by it; an option
// that fails in a constructor, which returns no error, is logged instead.
type OptionErrors struct {
	collecting bool
	errs       []error
}

// BeginOptions starts gathering option errors for a Configure call.
func (e *OptionErrors) BeginOptions() {
	e.collecting = true
	e.errs = nil
}

// EndOptions returns the errors gathered since BeginOptions, joined.
func (e *OptionErrors) EndOptions() error {
	err := errors.Join(e.errs...)
	e.collecting = false
	e.errs = nil
	return err
}

// Record keeps err for EndOptions and reports true, or reports false when
// no Configure call is gathering errors.
func (e *OptionErrors) Record(err error) bool {
	if e.collecting {
		e.errs = append(e.errs, err)
	}
	return e.collecting
}

// LogOptionError logs an option error that no Configure call will return.
func LogOptionError(app AppContext, path string, err error) {
	logger := slog.Default()
	if app != nil {
		logger = app.Logger()
	}
	logger.Warn("option not applied", "widget", path, "err", err)
}

// OptionFailed reports an option that could not be applied; the widget
// keeps its previous value. See OptionErrors.
func (b *Base) OptionFailed(err error) {
	if b.Record(err) {
		return
	}
	path := ""
	if b.Win != nil {
		path = b.Win.PathName
	}
	LogOptionError(b.App, path, err)
}

// LookupColor resolves a colour; on failure it calls OptionFailed and
// reports false so the caller keeps its previous value.
func (b *Base) LookupColor[C color.Spec](c C) (*color.Color, bool) {
	col, err := b.App.ColorCache().Resolve(c)
	if err != nil {
		b.OptionFailed(err)
		return nil, false
	}
	return col, true
}

// SetBackgroundColor sets -background and rebuilds the 3D border. It
// reports whether the colour was found.
func (b *Base) SetBackgroundColor[C color.Spec](c C) bool {
	col, ok := b.LookupColor(c)
	if !ok {
		return false
	}
	b.Background = col
	b.UpdateBorder()
	return true
}

// SetForegroundColor sets -foreground.
func (b *Base) SetForegroundColor[C color.Spec](c C) bool {
	col, ok := b.LookupColor(c)
	if ok {
		b.Foreground = col
	}
	return ok
}

// SetHighlightBackgroundColor sets -highlightbackground.
func (b *Base) SetHighlightBackgroundColor[C color.Spec](c C) bool {
	col, ok := b.LookupColor(c)
	if ok {
		b.HighlightBackground = col
	}
	return ok
}

// SetHighlightColor sets -highlightcolor.
func (b *Base) SetHighlightColor[C color.Spec](c C) bool {
	col, ok := b.LookupColor(c)
	if ok {
		b.HighlightColor = col
	}
	return ok
}

// SetFont sets -font from a font name, descriptor or attributes. It
// reports whether the font was found.
func (b *Base) SetFont[F font.Spec](f F) bool {
	fnt, err := b.App.FontRegistry().Resolve(f)
	if err != nil {
		b.OptionFailed(err)
		return false
	}
	b.Font = fnt
	return true
}

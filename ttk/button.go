package ttk

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Button is a themed button widget.
type Button struct {
	TtkWidget
	Text     string
	Command  func()
	Anchor   int // reserved
	Font     font.Font
	Img      widget.WidgetImage
	Compound widget.Compound
}

// GetText implements TextProvider.
func (b *Button) GetText() string { return b.Text }

// GetFont implements TextProvider.
func (b *Button) GetFont() font.Font { return b.Font }

// GetImage implements TextProvider.
func (b *Button) GetImage() widget.WidgetImage { return b.Img }

// GetCompound implements TextProvider.
func (b *Button) GetCompound() widget.Compound { return b.Compound }

// ButtonOption configures a Button.
type ButtonOption func(*Button)

// ButtonText sets the button text.
func ButtonText(s string) ButtonOption {
	return func(b *Button) { b.Text = s }
}

// ButtonCommand sets the callback invoked on click.
func ButtonCommand(fn func()) ButtonOption {
	return func(b *Button) { b.Command = fn }
}

// ButtonImage sets the image.
func ButtonImage(img widget.WidgetImage) ButtonOption {
	return func(b *Button) { b.Img = img }
}

// ButtonCompound sets how text and image are combined.
func ButtonCompound(c widget.Compound) ButtonOption {
	return func(b *Button) { b.Compound = c }
}

// ButtonFont sets the font.
func ButtonFont(name string) ButtonOption {
	return func(b *Button) {
		f, err := b.App.FontRegistry().Get(name)
		if err == nil {
			b.Font = f
		}
	}
}

// NewButton creates a themed button widget.
func NewButton(parent widget.Caregiver, name string, opts ...ButtonOption) *Button {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	b := &Button{}
	b.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	InitTtkWidget(&b.TtkWidget, win, app, "TButton")

	// Bind label element to this widget.
	if b.Theme != nil {
		labelFactory := NewLabelElementFactory(b)
		tmpl := b.Theme.GetLayout("TButton")
		if tmpl != nil {
			ctx := &DrawContext{
				Display: b.Context.Display,
				Depth:   b.Context.Depth,
				Style:   b.Context.Style,
			}
			b.Layout = newLayoutWithLabel(tmpl, b.Theme, ctx, b.Context.Style, labelFactory)
			b.Context = ctx
		}
	}

	for _, opt := range opts {
		opt(b)
	}

	// Recompute size.
	if b.Layout != nil {
		rw, rh := b.Layout.Size(b.State)
		if rw > 0 {
			win.ReqWidth = rw
		}
		if rh > 0 {
			win.ReqHeight = rh
		}
	}

	// Button-specific bindings.
	bindTtkButton(b, app)

	return b
}

// Invoke executes the button's command.
func (b *Button) Invoke() {
	if b.State&StateDisabled != 0 {
		return
	}
	if b.Command != nil {
		b.Command()
	}
}

// bindTtkButton adds press/release bindings for the themed button.
func bindTtkButton(b *Button, app widget.AppContext) {
	win := b.Win

	// Button1 press → +StatePressed.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if b.State&StateDisabled != 0 {
			return
		}
		if ev.Button == 1 {
			b.ChangeState(StatePressed, 0)
		}
	})

	// Button1 release → invoke if inside, -StatePressed.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if b.State&StateDisabled != 0 {
			return
		}
		if ev.Button == 1 {
			wasPressed := b.State&StatePressed != 0
			b.ChangeState(0, StatePressed)
			if wasPressed && ev.X >= 0 && ev.X < win.Width && ev.Y >= 0 && ev.Y < win.Height {
				b.Invoke()
			}
		}
	})
}

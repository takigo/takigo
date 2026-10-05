package ttk

import (
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// Button is a themed button widget.
type Button struct {
	TtkWidget
	Text      string
	Command   func()
	Anchor    int // reserved
	Font      font.Font
	Img       widget.WidgetImage
	Compound  widget.Compound
	Underline int
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
func ButtonFont[F font.Spec](name F) ButtonOption {
	return func(b *Button) {
		f, err := b.App.FontRegistry().Resolve(name)
		if err != nil {
			b.OptionFailed(err)
			return
		}
		b.Font = f
	}
}

// ButtonUnderline sets -underline: the index of the character to underline.
func ButtonUnderline(i int) ButtonOption {
	return func(b *Button) { b.Underline = i }
}

// GetUnderline implements the label element's optional underline provider.
func (b *Button) GetUnderline() int { return b.Underline }

// ButtonWidth sets -width: in average characters, a negative value being
// a minimum (as in the style's -width -9).
func ButtonWidth(n int) ButtonOption {
	return func(b *Button) { b.SetWidgetOption("-width", n) }
}

// ButtonStyleOpt overrides the TTK style name (e.g. "Toolbutton").
// Must be applied before other options that depend on the layout.
func ButtonStyleOpt(name string) ButtonOption {
	return func(b *Button) { b.StyleName = name }
}

// NewButton creates a themed button widget.
func NewButton(parent widget.Caregiver, name string, opts ...ButtonOption) *Button {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	b := &Button{Underline: -1}
	b.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	// Buttons are focusable via Tab traversal.
	win.Flags |= window.FlagFocusable

	initTtkBase(&b.TtkWidget, win, app, "TButton")
	b.reconfigure = func() { _ = b.Configure() }

	for _, opt := range opts {
		opt(b)
	}

	// Build the layout once, for the final style, with the label element
	// bound to this widget.
	if b.Theme != nil {
		if b.StyleName != "TButton" {
			b.setStyle(b.Theme.ResolveStyle(b.StyleName))
		}
		labelFactory := NewLabelElementFactory(b)
		b.LabelFactory = labelFactory
		tmpl := b.Theme.GetLayout(b.StyleName)
		if tmpl == nil {
			tmpl = b.Theme.GetLayout("TButton")
		}
		if tmpl != nil {
			ctx := &DrawContext{
				Display: b.Context.Display,
				Depth:   b.Context.Depth,
				Style:   b.Context.Style,
			}
			b.Layout = newLayoutWithLabel(tmpl, b.Theme, ctx, ctx.Style, labelFactory)
			b.Context = ctx
		}
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
	bindTtkHover(&b.TtkWidget, app)
	bindTtkButton(b, app)

	return b
}

// Configure sets options after creation.
func (b *Button) Configure(opts ...ButtonOption) error {
	return configure(&b.TtkWidget, b, opts, nil, nil)
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

	// Space key → invoke (matches Tk's "bind TButton <Key-space>" binding).
	app.Dispatcher().Bind(win.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_space {
			b.Invoke()
		}
	})
}

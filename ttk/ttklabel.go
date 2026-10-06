package ttk

import (
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// Label is a themed text/image display widget.
type Label struct {
	TtkWidget
	Text     string
	Anchor   int // reserved for future anchor support
	Font     font.Font
	Img      widget.WidgetImage
	Compound widget.Compound

	// TextVariable linkage — when set, the variable's value overrides Text.
	TextVar *widget.Variable[string]
	unsub   func()
}

// GetText implements TextProvider.
func (l *Label) GetText() string { return l.Text }

// GetFont implements TextProvider.
func (l *Label) GetFont() font.Font { return l.Font }

// GetImage implements TextProvider.
func (l *Label) GetImage() widget.WidgetImage { return l.Img }

// GetCompound implements TextProvider.
func (l *Label) GetCompound() widget.Compound { return l.Compound }

// LabelOption configures a Label.
type LabelOption func(*Label)

// LabelText sets the label text.
func LabelText(s string) LabelOption {
	return func(l *Label) { l.Text = s }
}

// LabelTextVariable links the label's text to a string variable.
// When the variable changes, the label text updates automatically.
func LabelTextVariable(v *widget.Variable[string]) LabelOption {
	return func(l *Label) {
		if l.unsub != nil {
			l.unsub()
		}
		l.TextVar = v
		l.Text = v.Get()
		l.unsub = v.OnChange(func(_, newVal string) {
			l.SetText(newVal)
		})
	}
}

// LabelForeground sets the text color (pixel value) in the style.
func LabelForeground(pixel uint64) LabelOption {
	return func(l *Label) { l.SetWidgetOption("-foreground", pixel) }
}

// LabelFont sets the font.
func LabelFont[F font.Spec](name F) LabelOption {
	return func(l *Label) {
		f, err := l.App.FontRegistry().Resolve(name)
		if err != nil {
			l.OptionFailed(err)
			return
		}
		l.Font = f
	}
}

// LabelImage sets the image.
func LabelImage(img widget.WidgetImage) LabelOption {
	return func(l *Label) { l.Img = img }
}

// LabelWrapLength sets -wraplength (a Tk distance such as "4i").
func LabelWrapLength[L screenunit.Length](v L) LabelOption {
	return func(l *Label) { l.SetWidgetOption("-wraplength", screenunit.ToPixels(v)) }
}

// LabelJustify sets -justify for multi-line text.
func LabelJustify(j option.Justify) LabelOption {
	return func(l *Label) { l.SetWidgetOption("-justify", j) }
}

// LabelAnchor sets -anchor: where the content sits in the label.
func LabelAnchor(a option.Anchor) LabelOption {
	return func(l *Label) { l.SetWidgetOption("-anchor", a) }
}

// LabelPadding sets -padding as a Tk padding spec ("10 2 10 6", "2.25p").
func LabelPadding(spec string) LabelOption {
	return func(l *Label) { l.SetWidgetOption("-padding", spec) }
}

// LabelCompound sets how text and image are combined.
func LabelCompound(c widget.Compound) LabelOption {
	return func(l *Label) { l.Compound = c }
}

// NewLabel creates a themed label widget.
func NewLabel(parent widget.Caregiver, name string, opts ...LabelOption) *Label {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	l := &Label{}

	// Get default font.
	l.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	// Initialize TTK base — this needs to happen before options that
	// reference the label element, but the label element needs the widget
	// as TextProvider. We solve this with a two-pass approach:
	// 1. Create the widget and init TTK base with a temporary layout.
	// 2. Register the label element factory with this widget, re-create layout.
	InitTtkWidget(&l.TtkWidget, win, app, "TLabel")
	l.reconfigure = func() { _ = l.Configure() }

	// Register a label element factory bound to this specific widget.
	if l.Theme != nil {
		labelFactory := NewLabelElementFactory(l)
		l.LabelFactory = labelFactory
		// Recreate layout with bound label element.
		tmpl := l.Theme.GetLayout("TLabel")
		if tmpl != nil {
			ctx := &DrawContext{
				Display: l.Context.Display,
				Depth:   l.Context.Depth,
				Style:   l.Context.Style,
			}
			l.Layout = newLayoutWithLabel(tmpl, l.Theme, ctx, l.Context.Style, labelFactory)
			l.Context = ctx
		}
	}

	for _, opt := range opts {
		opt(l)
	}
	win.OnDestroy(func() {
		if l.unsub != nil {
			l.unsub()
			l.unsub = nil
		}
	})

	l.updateReqSize()
	return l
}

// SetText changes the text; like Tk's -text/-textvariable traces it
// re-requests the widget size (TtkResizeWidget) before redrawing.
func (l *Label) SetText(s string) {
	l.Text = s
	l.resize(l.updateReqSize)
	l.Display()
}

// Configure sets options after creation.
func (l *Label) Configure(opts ...LabelOption) error {
	return configure(&l.TtkWidget, l, opts, nil, l.updateReqSize)
}

func (l *Label) updateReqSize() {
	if l.Layout == nil {
		return
	}
	rw, rh := l.Layout.Size(l.State)
	if rw > 0 {
		l.Win.ReqWidth = rw
	}
	if rh > 0 {
		l.Win.ReqHeight = rh
	}
}

// newLayoutWithLabel instantiates a layout, overriding the "label" element with a bound factory.
func newLayoutWithLabel(tmpl *LayoutTemplate, theme *Theme, ctx *DrawContext, style *Style, labelFactory ElementFactory) *Layout {
	l := &Layout{Style: style, Context: ctx}
	l.Root = instantiateNodeWithLabel(tmpl, theme, ctx, labelFactory)
	return l
}

func instantiateNodeWithLabel(tmpl *LayoutTemplate, theme *Theme, ctx *DrawContext, labelFactory ElementFactory) *LayoutNode {
	return instantiateNodeWithLabelDepth(tmpl, theme, ctx, labelFactory, 0)
}

func instantiateNodeWithLabelDepth(tmpl *LayoutTemplate, theme *Theme, ctx *DrawContext, labelFactory ElementFactory, depth int) *LayoutNode {
	if tmpl == nil || depth > 1000 {
		return nil
	}
	var elem Element
	if tmpl.ElementName == "label" {
		elem = labelFactory(ctx)
	} else {
		factory := theme.GetElement(tmpl.ElementName)
		elem = factory(ctx)
	}
	node := &LayoutNode{
		Name:    tmpl.ElementName,
		Flags:   tmpl.Flags,
		Element: elem,
	}
	node.Children = instantiateNodeWithLabelDepth(tmpl.Children, theme, ctx, labelFactory, depth+1)
	node.Next = instantiateNodeWithLabelDepth(tmpl.Next, theme, ctx, labelFactory, depth)
	return node
}

package ttk

import (
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

// Element is the interface for TTK drawing elements.
// Elements know their intrinsic size and how to draw themselves.
type Element interface {
	// Size returns the element's intrinsic width, height, and internal padding.
	Size(state State) (w, h int, padding Padding)
	// Draw draws the element into the given box.
	Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State)
}

// ElementFactory creates an Element bound to a DrawContext.
type ElementFactory func(ctx *DrawContext) Element

// DrawContext provides shared resources to elements during creation and drawing.
type DrawContext struct {
	Display platform.DisplayServer
	Depth   int
	Style   *Style
}

// TextProvider lets the label element read widget text without import cycles.
type TextProvider interface {
	GetText() string
	GetFont() font.Font
	GetImage() widget.WidgetImage
	GetCompound() widget.Compound
}

// NullElement is a zero-size element that draws nothing.
type NullElement struct{}

func (NullElement) Size(State) (int, int, Padding) { return 0, 0, Padding{} }
func (NullElement) Draw(platform.DisplayServer, platform.DrawableID, platform.GCID, Box, State) {}

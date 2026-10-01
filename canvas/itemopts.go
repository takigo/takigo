package canvas

import (
	imgcolor "image/color"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
)

// ItemOption is a functional option for configuring canvas items.
type ItemOption func(canvas *Canvas, item Item) error

// FillColor sets the fill color of an item.
func FillColor[C color.Spec](name C) ItemOption {
	return func(c *Canvas, item Item) error {
		col, err := c.ColorCache().Resolve(name)
		if err != nil {
			return err
		}
		ref := col.Ref()
		switch it := item.(type) {
		case *RectOvalItem:
			it.fill = ref
		case *PolygonItem:
			it.fill = ref
		case *ArcItem:
			it.fill = ref
		case *TextItem: // -fill is the text colour
			it.color = ref
		case *LineItem: // -fill is the line colour
			it.color = ref
		}
		return nil
	}
}

// OutlineColor sets the outline color of an item.
func OutlineColor[C color.Spec](name C) ItemOption {
	return func(c *Canvas, item Item) error {
		col, err := c.ColorCache().Resolve(name)
		if err != nil {
			return err
		}
		ref := col.Ref()
		switch it := item.(type) {
		case *RectOvalItem:
			it.outline = ref
		case *LineItem:
			it.color = ref
		case *PolygonItem:
			it.outline = ref
		case *ArcItem:
			it.outline = ref
		}
		return nil
	}
}

// OutlineWidth sets the outline/line width.
func OutlineWidth(w int) ItemOption {
	return func(_ *Canvas, item Item) error {
		switch it := item.(type) {
		case *RectOvalItem:
			it.outlineWidth = w
		case *LineItem:
			it.width = w
		case *PolygonItem:
			it.outlineWidth = w
		case *ArcItem:
			it.outlineWidth = w
		}
		return nil
	}
}

// Tags sets the tags on an item.
func Tags(tags ...string) ItemOption {
	return func(_ *Canvas, item Item) error {
		base := itemBase(item)
		if base == nil {
			return nil
		}
		c := base.canvas
		indexed := c != nil && c.idMap[base.ID] != nil
		if indexed {
			for _, t := range base.Tags {
				c.tagIndexRemove(t, base.ID)
			}
		}
		base.Tags = append([]string{}, tags...)
		if indexed {
			for _, t := range base.Tags {
				c.tagIndexAdd(t, base.ID)
			}
		}
		return nil
	}
}

// Dash sets the dash pattern for outlines/lines.
func Dash(pattern ...byte) ItemOption {
	return func(_ *Canvas, item Item) error {
		switch it := item.(type) {
		case *RectOvalItem:
			it.dash = pattern
		case *LineItem:
			it.dash = pattern
		case *PolygonItem:
			it.dash = pattern
		case *ArcItem:
			it.dash = pattern
		}
		return nil
	}
}

// AnchorOpt sets the anchor for text/image items.
func AnchorOpt(a option.Anchor) ItemOption {
	return func(_ *Canvas, item Item) error {
		switch it := item.(type) {
		case *TextItem:
			it.anchor = a
		case *ImageItem:
			it.anchor = a
		case *BitmapItem:
			it.anchor = a
		case *WindowItem:
			it.anchor = a
		}
		return nil
	}
}

// BitmapForeground sets the foreground (bit=1) color for a BitmapItem.
func BitmapForeground(r, g, b uint8) ItemOption {
	return func(_ *Canvas, item Item) error {
		if bi, ok := item.(*BitmapItem); ok {
			bi.Foreground = imgcolor.RGBA{R: r, G: g, B: b, A: 255}
		}
		return nil
	}
}

// BitmapBackground sets the background (bit=0) color for a BitmapItem.
// Use A=0 for transparent (default).
func BitmapBackground(r, g, b, a uint8) ItemOption {
	return func(_ *Canvas, item Item) error {
		if bi, ok := item.(*BitmapItem); ok {
			bi.Background = imgcolor.RGBA{R: r, G: g, B: b, A: a}
		}
		return nil
	}
}

// FontOpt sets the font for text items.
func FontOpt[F font.Spec](name F) ItemOption {
	return func(c *Canvas, item Item) error {
		f, err := c.FontRegistry().Resolve(name)
		if err != nil {
			return err
		}
		if it, ok := item.(*TextItem); ok {
			it.font = f
		}
		return nil
	}
}

// TextOpt sets the text string for text items.
func TextOpt(s string) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*TextItem); ok {
			it.text = s
			it.updateBBox()
		}
		return nil
	}
}

// TextAngle sets the rotation angle in degrees (clockwise on screen) for text items.
func TextAngle(deg float64) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*TextItem); ok {
			it.angle = deg
			it.updateBBox()
		}
		return nil
	}
}

// TextColor sets the color for text items.
func TextColor[C color.Spec](name C) ItemOption {
	return func(c *Canvas, item Item) error {
		col, err := c.ColorCache().Resolve(name)
		if err != nil {
			return err
		}
		if it, ok := item.(*TextItem); ok {
			it.color = col.Ref()
		}
		return nil
	}
}

// ImageOpt sets the image for image items.
func ImageOpt(img widget.WidgetImage) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*ImageItem); ok {
			it.image = img
			it.updateBBox()
		}
		return nil
	}
}

// StartAngle sets the start angle (degrees) for arc items.
func StartAngle(deg float64) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*ArcItem); ok {
			it.start = deg
		}
		return nil
	}
}

// Extent sets the angular extent (degrees) for arc items.
func Extent(deg float64) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*ArcItem); ok {
			it.extent = deg
		}
		return nil
	}
}

// ArcStyle is the drawing style for arc items.
type ArcStyle int

const (
	ArcStyleArc      ArcStyle = iota // stroke only
	ArcStyleChord                    // fill chord
	ArcStylePieslice                 // fill pieslice
)

// ArcStyleOpt sets the style for arc items.
func ArcStyleOpt(s ArcStyle) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*ArcItem); ok {
			it.style = s
		}
		return nil
	}
}

// ArrowMode specifies which ends of a line have arrows.
type ArrowMode int

const (
	ArrowNone ArrowMode = iota
	ArrowFirst
	ArrowLast
	ArrowBoth
)

// Arrow sets the arrow mode for line items.
func Arrow(mode ArrowMode) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*LineItem); ok {
			it.arrow = mode
		}
		return nil
	}
}

// ArrowShape sets the arrow dimensions (a=along shaft, b=along side, c=base width).
func ArrowShape(a, b, c float64) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*LineItem); ok {
			it.arrowShapeA = a
			it.arrowShapeB = b
			it.arrowShapeC = c
		}
		return nil
	}
}

// Smooth enables Bezier spline smoothing.
func Smooth(on bool) ItemOption {
	return func(_ *Canvas, item Item) error {
		switch it := item.(type) {
		case *LineItem:
			it.smooth = on
		case *PolygonItem:
			it.smooth = on
		}
		return nil
	}
}

// SplineSteps sets the number of steps for Bezier smoothing.
func SplineSteps(n int) ItemOption {
	return func(_ *Canvas, item Item) error {
		switch it := item.(type) {
		case *LineItem:
			it.splineSteps = n
		case *PolygonItem:
			it.splineSteps = n
		}
		return nil
	}
}

// CapStyleOpt sets the cap style for line items.
func CapStyleOpt(cap int) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*LineItem); ok {
			it.capStyle = cap
		}
		return nil
	}
}

// JoinStyleOpt sets the join style for line items.
func JoinStyleOpt(join int) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*LineItem); ok {
			it.joinStyle = join
		}
		return nil
	}
}

// JustifyOpt sets the text justification for text items.
func JustifyOpt(j option.Justify) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*TextItem); ok {
			it.justify = j
		}
		return nil
	}
}

// WidthOpt sets the wrap length for text items.
func WidthOpt(w int) ItemOption {
	return func(_ *Canvas, item Item) error {
		if it, ok := item.(*TextItem); ok {
			it.wrapLength = w
			it.updateBBox()
		}
		return nil
	}
}

// OutlineNone removes the outline from an item (no border drawn).
func OutlineNone() ItemOption {
	return func(_ *Canvas, item Item) error {
		switch it := item.(type) {
		case *RectOvalItem:
			it.outline = nil
		case *PolygonItem:
			it.outline = nil
		case *ArcItem:
			it.outline = nil
		}
		return nil
	}
}

// FillNone removes the fill from an item (transparent interior).
func FillNone() ItemOption {
	return func(_ *Canvas, item Item) error {
		switch it := item.(type) {
		case *RectOvalItem:
			it.fill = nil
		case *PolygonItem:
			it.fill = nil
		case *ArcItem:
			it.fill = nil
		}
		return nil
	}
}

// StateOpt sets the item state.
func StateOpt(s ItemState) ItemOption {
	return func(_ *Canvas, item Item) error {
		if base := itemBase(item); base != nil {
			base.state = s
		}
		return nil
	}
}

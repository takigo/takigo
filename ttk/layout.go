package ttk

import "github.com/msorc/takigo/platform"

// Position flags for layout elements. Ported from ttkLayout.c. As with
// TTK_PACK_*, a node without a Pack flag is given the whole cavity and does
// not consume it.
const (
	packSet    = 0x40
	PackTop    = packSet | 0x0
	PackBottom = packSet | 0x1
	PackLeft   = packSet | 0x2
	PackRight  = packSet | 0x3
	_packMask  = 0x3

	Expand = 0x4  // expand to fill available space
	Border = 0x8  // draw after children (border elements)
	FillXF = 0x10 // fill horizontally
	FillYF = 0x20 // fill vertically
	FillF  = FillXF | FillYF
)

// LayoutTemplate is a declarative tree definition for widget layouts.
type LayoutTemplate struct {
	ElementName string
	Flags       uint
	Children    *LayoutTemplate
	Next        *LayoutTemplate
}

// L builds a LayoutTemplate node concisely.
func L(name string, flags uint, children ...*LayoutTemplate) *LayoutTemplate {
	t := &LayoutTemplate{ElementName: name, Flags: flags}
	if len(children) > 0 {
		t.Children = children[0]
		for i := 1; i < len(children); i++ {
			children[i-1].Next = children[i]
		}
	}
	return t
}

// LayoutNode is an instantiated layout element with a bound element and parcel.
type LayoutNode struct {
	Name     string
	Flags    uint
	Element  Element
	Parcel   Box
	Children *LayoutNode
	Next     *LayoutNode
}

// Layout holds the root of an instantiated layout tree.
type Layout struct {
	Root    *LayoutNode
	Style   *Style
	Context *DrawContext
}

// NewLayout instantiates a layout tree from a template, binding elements via the theme.
func NewLayout(tmpl *LayoutTemplate, theme *Theme, ctx *DrawContext, style *Style) *Layout {
	l := &Layout{Style: style, Context: ctx}
	l.Root = instantiateNode(tmpl, theme, ctx)
	return l
}

func instantiateNode(tmpl *LayoutTemplate, theme *Theme, ctx *DrawContext) *LayoutNode {
	if tmpl == nil {
		return nil
	}
	factory := theme.GetElement(tmpl.ElementName)
	node := &LayoutNode{
		Name:    tmpl.ElementName,
		Flags:   tmpl.Flags,
		Element: factory(ctx),
	}
	node.Children = instantiateNode(tmpl.Children, theme, ctx)
	node.Next = instantiateNode(tmpl.Next, theme, ctx)
	return node
}

// Size computes the total requested size of the layout tree.
func (l *Layout) Size(state State) (int, int) {
	if l.Root == nil {
		return 0, 0
	}
	return nodeSize(l.Root, state, 0)
}

func nodeSize(n *LayoutNode, state State, depth int) (int, int) {
	if n == nil || depth > 1000 {
		return 0, 0
	}

	// Element's own size and padding.
	ew, eh, epad := n.Element.Size(state)
	_ = ew
	_ = eh

	// Children size (packed sequentially).
	cw, ch := childrenSize(n.Children, state, depth+1)

	// Content size = max of element content and children.
	w := max(ew, cw) + epad.Width()
	h := max(eh, ch) + epad.Height()

	// Accumulate siblings (Ttk_NodeListSize).
	if n.Next != nil {
		nw, nh := nodeSize(n.Next, state, depth)
		switch {
		case n.Flags&packSet == 0:
			w, h = max(w, nw), max(h, nh)
		case Side(n.Flags&_packMask) == SideTop || Side(n.Flags&_packMask) == SideBottom:
			w = max(w, nw)
			h += nh
		default:
			w += nw
			h = max(h, nh)
		}
	}

	return w, h
}

func childrenSize(n *LayoutNode, state State, depth int) (int, int) {
	if n == nil || depth > 1000 {
		return 0, 0
	}
	w, h := nodeSize(n, state, depth)
	return w, h
}

// Place assigns parcels to all nodes using cavity-based packing.
func (l *Layout) Place(state State, bounds Box) {
	if l.Root == nil {
		return
	}
	placeNodes(l.Root, state, bounds, 0)
}

func placeNodes(n *LayoutNode, state State, cavity Box, depth int) {
	for cur := n; cur != nil; cur = cur.Next {
		ew, eh, epad := cur.Element.Size(state)

		// Children size for this node.
		cw, ch := childrenSize(cur.Children, state, depth+1)
		nodeW := max(ew, cw) + epad.Width()
		nodeH := max(eh, ch) + epad.Height()

		// Determine pack side.
		side := Side(cur.Flags & _packMask)

		// Requested size — expand if flagged.
		reqW, reqH := nodeW, nodeH
		if cur.Flags&Expand != 0 {
			switch side {
			case SideTop, SideBottom:
				reqW = cavity.Width
				reqH = max(reqH, cavity.Height)
			case SideLeft, SideRight:
				reqH = cavity.Height
				reqW = max(reqW, cavity.Width)
			}
		}

		// Carve parcel from cavity (Ttk_PositionBox).
		parcel := cavity
		if cur.Flags&packSet != 0 {
			parcel = PackBox(&cavity, reqW, reqH, side)
		}

		// Apply sticky.
		sticky := Sticky(0)
		if cur.Flags&FillXF != 0 {
			sticky |= FillX
		}
		if cur.Flags&FillYF != 0 {
			sticky |= FillY
		}
		if sticky != 0 || cur.Flags&Expand != 0 {
			sticky |= FillBoth
		}
		cur.Parcel = StickBox(parcel, nodeW, nodeH, sticky)

		// Place children inside the element's inner area.
		if cur.Children != nil {
			inner := PadBox(cur.Parcel, epad)
			placeNodes(cur.Children, state, inner, depth+1)
		}
	}
}

// Draw renders all elements in the layout tree.
func (l *Layout) Draw(state State, d DrawArgs) {
	if l.Root == nil {
		return
	}
	// Two passes: non-border first, then border elements.
	drawNodes(l.Root, state, d, false)
	drawNodes(l.Root, state, d, true)
}

// DrawArgs bundles drawing parameters.
type DrawArgs struct {
	Display  platform.DisplayServer
	Drawable platform.DrawableID
	GC       platform.GCID
}

func drawNodes(n *LayoutNode, state State, d DrawArgs, borderPass bool) {
	for cur := n; cur != nil; cur = cur.Next {
		isBorder := cur.Flags&Border != 0
		if isBorder == borderPass {
			cur.Element.Draw(d.Display, d.Drawable, d.GC, cur.Parcel, state)
		}
		if cur.Children != nil {
			drawNodes(cur.Children, state, d, borderPass)
		}
	}
}

// ClientRegion returns the inner box of the named element (parcel minus padding).
func (l *Layout) ClientRegion(name string, state State) Box {
	node := findNode(l.Root, name)
	if node == nil {
		return Box{}
	}
	_, _, epad := node.Element.Size(state)
	return PadBox(node.Parcel, epad)
}

func findNode(n *LayoutNode, name string) *LayoutNode {
	for cur := n; cur != nil; cur = cur.Next {
		if cur.Name == name {
			return cur
		}
		if cur.Children != nil {
			if found := findNode(cur.Children, name); found != nil {
				return found
			}
		}
	}
	return nil
}

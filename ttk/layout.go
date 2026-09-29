package ttk

import "github.com/msorc/takigo/platform"

// Position flags for layout elements. Ported from ttkLayout.c. As with
// TTK_PACK_*, a node without a Pack flag is given the whole cavity and does
// not consume it.
const (
	packSet    = 0x40
	PackTop    = packSet
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

	// Set by measure: the node's own size (element and children, without
	// its siblings) and the element's padding.
	w, h int
	pad  Padding
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
	return instantiateNodeDepth(tmpl, theme, ctx, 0)
}

func instantiateNodeDepth(tmpl *LayoutTemplate, theme *Theme, ctx *DrawContext, depth int) *LayoutNode {
	if tmpl == nil || depth > 1000 {
		return nil
	}
	factory := theme.GetElement(tmpl.ElementName)
	node := &LayoutNode{
		Name:    tmpl.ElementName,
		Flags:   tmpl.Flags,
		Element: factory(ctx),
	}
	node.Children = instantiateNodeDepth(tmpl.Children, theme, ctx, depth+1)
	node.Next = instantiateNodeDepth(tmpl.Next, theme, ctx, depth)
	return node
}

// Size computes the total requested size of the layout tree.
func (l *Layout) Size(state State) (int, int) {
	if l.Root == nil {
		return 0, 0
	}
	return measure(l.Root, state, 0)
}

// measure sizes n, its descendants and its following siblings bottom-up,
// asking each element for its size once and recording every node's own
// size, and returns the size of the sibling list that starts at n
// (Ttk_NodeListSize). Place reuses the recorded sizes instead of sizing
// each subtree again at every ancestor level.
func measure(n *LayoutNode, state State, depth int) (int, int) {
	if n == nil || depth > 1000 {
		return 0, 0
	}
	ew, eh, epad := n.Element.Size(state)
	cw, ch := measure(n.Children, state, depth+1)
	n.w = max(ew, cw) + epad.Width()
	n.h = max(eh, ch) + epad.Height()
	n.pad = epad

	w, h := n.w, n.h
	if n.Next != nil {
		nw, nh := measure(n.Next, state, depth)
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

// Place assigns parcels to all nodes using cavity-based packing.
func (l *Layout) Place(state State, bounds Box) {
	if l.Root == nil {
		return
	}
	measure(l.Root, state, 0)
	placeNodes(l.Root, state, bounds, 0)
}

func placeNodes(n *LayoutNode, state State, cavity Box, depth int) {
	for cur := n; cur != nil; cur = cur.Next {
		nodeW, nodeH, epad := cur.w, cur.h, cur.pad

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
	drawNodesDepth(n, state, d, borderPass, 0)
}

func drawNodesDepth(n *LayoutNode, state State, d DrawArgs, borderPass bool, depth int) {
	if n == nil || depth > 1000 {
		return
	}
	for cur := n; cur != nil; cur = cur.Next {
		isBorder := cur.Flags&Border != 0
		if isBorder == borderPass {
			cur.Element.Draw(d.Display, d.Drawable, d.GC, cur.Parcel, state)
		}
		if cur.Children != nil {
			drawNodesDepth(cur.Children, state, d, borderPass, depth+1)
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
	return findNodeDepth(n, name, 0)
}

func findNodeDepth(n *LayoutNode, name string, depth int) *LayoutNode {
	if n == nil || depth > 1000 {
		return nil
	}
	for cur := n; cur != nil; cur = cur.Next {
		if cur.Name == name {
			return cur
		}
		if cur.Children != nil {
			if found := findNodeDepth(cur.Children, name, depth+1); found != nil {
				return found
			}
		}
	}
	return nil
}

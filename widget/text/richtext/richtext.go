// Package richtext shows HTML in a text widget: a subset of HTML styled by
// a subset of CSS, turned into the widget's tags, embedded images and
// bindings. It is meant for read-only documents such as help pages, about
// boxes and release notes, not for browsing the web.
//
// Supported elements: headings, p, div and the HTML5 sectioning elements,
// br, hr, pre, blockquote, ul/ol/li (nested, with start, value and type),
// dl/dt/dd, a (links and #anchors), img, table/tr/td/th (inline cell
// content, wrapped to columns laid out on tab stops), b/strong, i/em, u, s/del,
// code/kbd/samp/tt, sub/sup, small/big, mark, span, font and center.
//
// CSS comes from DefaultStylesheet, the document's <style> elements, the
// Stylesheet option and style attributes, cascaded by specificity. Type,
// class, id, :hover, :link and :first-child selectors combine with the
// descendant and child combinators. The properties are color, background,
// font (-family, -size, -weight, -style), text-decoration, text-align,
// text-indent, white-space, vertical-align, list-style(-type), display,
// margin, padding, border, border-radius and, for images, width, height,
// max-width and float (left or right, text.ImageFloatOpt), and clear on
// blocks.
//
// A block with a background, border or padding becomes a block box
// (text.TagBlockBackground and the like). Margins collapse between blocks
// as in CSS.
package richtext

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/image"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/text"
	"github.com/takigo/takigo/window"
)

// tagPrefix starts the names of the tags this package creates.
const tagPrefix = "richtext:"

// Option configures Render and Insert.
type Option func(*config)

type config struct {
	css       []string
	resolver  func(src string) (widget.WidgetImage, error)
	baseDir   string
	onLink    func(href string)
	baseFont  *font.Attributes
	fitImages bool
	reflow    bool
	width     int
	measure   func(cs charStyle, s string) int
	images    map[string]widget.WidgetImage
}

// Stylesheet adds author CSS, applied after the document's own <style>
// elements.
func Stylesheet(css string) Option {
	return func(c *config) { c.css = append(c.css, css) }
}

// ImageResolver loads the image of an <img src>. It replaces the default
// loader, which reads data: URIs and files (relative to BaseDir) but never
// the network. Return nil or an error to show the alt text.
func ImageResolver(fn func(src string) (widget.WidgetImage, error)) Option {
	return func(c *config) { c.resolver = fn }
}

// BaseDir is the directory relative image paths are read from.
func BaseDir(dir string) Option {
	return func(c *config) { c.baseDir = dir }
}

// OnLink is called with the href of a clicked link. A link to "#name"
// scrolls to the element with that id (or <a name>) instead when the
// document has one.
func OnLink(fn func(href string)) Option {
	return func(c *config) { c.onLink = fn }
}

// BaseFont is the font of the document's text, from which ems and the
// heading sizes follow; TkDefaultFont by default.
func BaseFont(attrs font.Attributes) Option {
	return func(c *config) { c.baseFont = &attrs }
}

// FitImages scales images down to the width available when they are wider
// (on by default).
func FitImages(on bool) Option {
	return func(c *config) { c.fitImages = on }
}

// Reflow re-renders the document when the widget's width changes (on by
// default), so tables and fitted images follow the window; the view stays
// near the same place.
func Reflow(on bool) Option {
	return func(c *config) { c.reflow = on }
}

var generation atomic.Uint64

// reflowKey holds a widget's reflowState.
var reflowKey = new(window.ValueKey)

// reflowState is what a widget needs to render its document again.
type reflowState struct {
	html    string
	opts    []Option
	width   int
	pending bool
	active  bool
}

// watchWidth renders the document again at idle time when the widget's
// text area changes width.
func watchWidth(tw *text.TextWidget, html string, opts []Option, width int) {
	w := tw.Window()
	st, _ := w.Value(reflowKey).(*reflowState)
	if st == nil {
		st = &reflowState{}
		w.SetValue(reflowKey, st)
		w.OnConfigure(func() {
			if !st.active || st.pending || textWidth(tw) == st.width || tw.Destroyed() {
				return
			}
			st.pending = true
			tw.App.DoWhenIdle(func() {
				st.pending = false
				if tw.Destroyed() || !st.active || textWidth(tw) == st.width {
					return
				}
				frac, _ := tw.YViewFractions()
				_ = Render(tw, st.html, st.opts...)
				tw.YViewMoveTo(frac)
			})
		})
	}
	st.html, st.opts, st.width, st.active = html, opts, width, true
}

// Render replaces the widget's content with the HTML document. It also
// sets the widget's font to the base font and word wrapping, and scrolls
// to the top. Render does not record undo.
func Render(tw *text.TextWidget, html string, opts ...Option) error {
	Clear(tw)
	cfg := newConfig(tw, opts)
	if attrs := cfg.base(tw); attrs.Family != "" || attrs.Size != 0 {
		_ = tw.Configure(text.FontOpt(attrs), text.WrapModeOpt(text.WrapWord))
	} else {
		_ = tw.Configure(text.WrapModeOpt(text.WrapWord))
	}
	cfg.width = textWidth(tw)
	err := insert(tw, text.Index{Line: 1}, html, cfg)
	tw.Doc().UndoStack().Reset()
	tw.YView(1)
	if cfg.reflow {
		watchWidth(tw, html, opts, cfg.width)
	} else if st, _ := tw.Window().Value(reflowKey).(*reflowState); st != nil {
		st.active = false
	}
	return err
}

// Insert adds the HTML at index, keeping the rest of the content.
func Insert[I text.IndexSpec](tw *text.TextWidget, index I, html string, opts ...Option) error {
	idx, err := tw.Index(index)
	if err != nil {
		return err
	}
	return insert(tw, idx, html, newConfig(tw, opts))
}

// Clear deletes the widget's content and every tag Render and Insert made,
// and stops reflowing it.
func Clear(tw *text.TextWidget) {
	if st, _ := tw.Window().Value(reflowKey).(*reflowState); st != nil {
		st.active = false
	}
	var names []string
	for _, n := range tw.TagNames() {
		if strings.HasPrefix(n, tagPrefix) {
			names = append(names, n)
		}
	}
	tw.TagDelete(names...)
	doc := tw.Doc()
	doc.Delete(text.Index{Line: 1}, doc.EndIndex())
	for _, m := range doc.MarkNames() {
		if strings.HasPrefix(m, tagPrefix) {
			doc.MarkUnset(m)
		}
	}
	tw.Display()
}

func newConfig(tw *text.TextWidget, opts []Option) *config {
	cfg := &config{fitImages: true, reflow: true, images: map[string]widget.WidgetImage{}}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.width = textWidth(tw)
	reg := tw.App.FontRegistry()
	cfg.measure = func(cs charStyle, s string) int {
		f, err := reg.Resolve(cs.attrs())
		if err != nil {
			return 0
		}
		return f.MeasureString(s)
	}
	return cfg
}

// textWidth is the width of the widget's text area: of its window, or
// the width it asks for before it is laid out.
func textWidth(tw *text.TextWidget) int {
	w := tw.Window().Width
	if w <= 1 {
		w = tw.Window().ReqWidth
	}
	return max(w-2*(tw.BorderWidth+tw.HighlightWidth+tw.PadX), 0)
}

// base returns the base font's attributes.
func (cfg *config) base(tw *text.TextWidget) font.Attributes {
	if cfg.baseFont != nil {
		return *cfg.baseFont
	}
	if f, err := tw.App.FontRegistry().Get(font.TkDefaultFont); err == nil {
		return f.Attrs()
	}
	return font.Attributes{}
}

func (cs charStyle) attrs() font.Attributes {
	a := font.Attributes{Family: cs.family, Size: -float64(cs.size)}
	if cs.bold {
		a.Weight = font.WeightBold
	}
	if cs.italic {
		a.Slant = font.SlantItalic
	}
	return a
}

// newStyler sets up the cascade and the root style for a base font.
func newStyler(base font.Attributes, docCSS []string, cfg *config) (*styler, style) {
	ptPx := screenunit.Pt(1).Float()
	scale := max(screenunit.DPI()/96, 1)
	size := 13 * scale
	switch {
	case base.Size > 0:
		size = base.Size * ptPx
	case base.Size < 0:
		size = -base.Size
	}
	st := &styler{rootPx: size, scale: scale, ptPx: ptPx}
	st.sheet.parseCSS(DefaultStylesheet, 0)
	for _, css := range docCSS {
		st.sheet.parseCSS(css, 1)
	}
	for _, css := range cfg.css {
		st.sheet.parseCSS(css, 1)
	}
	root := style{family: base.Family, size: size, bold: base.Weight == font.WeightBold,
		italic: base.Slant == font.SlantItalic, whiteSpace: "normal", listStyle: "disc", display: "block"}
	return st, root
}

// build parses the document and lays it out as a plan.
func build(html string, base font.Attributes, cfg *config, prefix string) *plan {
	return &buildDoc(html, base, cfg, prefix).p
}

func buildDoc(html string, base font.Attributes, cfg *config, prefix string) *builder {
	tree, docCSS := parseHTML(html)
	st, root := newStyler(base, docCSS, cfg)
	if !hasElement(tree, "body") {
		tree.tag = "body"
	}
	b := newBuilder(cfg, st, prefix)
	b.blockStyle = b.charStyleOf(&root, inline{})
	b.lineNL = b.blockStyle
	if tree.tag == "body" {
		b.walk(tree, &root, inline{})
	} else {
		for _, c := range tree.children {
			b.walk(c, &root, inline{})
		}
	}
	b.finish()
	slices.SortStableFunc(b.p.tags, func(x, y tagDef) int { return x.kind - y.kind })
	return b
}

func hasElement(n *node, tag string) bool {
	if n.kind == elementNode && n.tag == tag {
		return true
	}
	for _, c := range n.children {
		if c.kind == elementNode && hasElement(c, tag) {
			return true
		}
	}
	return false
}

// insert builds the plan and applies it at idx.
func insert(tw *text.TextWidget, idx text.Index, html string, cfg *config) error {
	prefix := fmt.Sprintf("%s%d:", tagPrefix, generation.Add(1))
	p := build(html, cfg.base(tw), cfg, prefix)
	for _, td := range p.tags {
		tw.TagConfigure(td.name, td.opts...)
	}
	doc := tw.Doc()
	at := idx
	pieces := strings.Split(p.text.String(), "\uFFFC")
	for i, piece := range pieces {
		if piece != "" {
			at = doc.Insert(at, piece)
		}
		if i < len(p.images) {
			im := p.images[i]
			if err := tw.ImageCreate(at, im.img, im.opts...); err != nil {
				return err
			}
			at.Char++
		}
	}
	conv := func(q pos) text.Index {
		if q.line == 0 {
			return text.Index{Line: idx.Line, Char: idx.Char + q.char}
		}
		return text.Index{Line: idx.Line + q.line, Char: q.char}
	}
	for _, r := range p.ranges {
		doc.TagAdd(r.name, conv(r.start), conv(r.end))
	}
	for name, q := range p.anchors {
		doc.MarkSet(prefix+"#"+name, conv(q))
		doc.Marks[prefix+"#"+name].Gravity = text.GravityLeft
	}
	for _, l := range p.links {
		bindLink(tw, l, prefix, cfg.onLink)
	}
	tw.Display()
	return nil
}

func bindLink(tw *text.TextWidget, l link, prefix string, onLink func(string)) {
	href := l.href
	tw.TagBind(l.tag, "<Button-1>", func() {
		if name, ok := strings.CutPrefix(href, "#"); ok {
			if m := tw.Doc().MarkPos(prefix + "#" + name); m != nil {
				tw.YView(m.Line)
				return
			}
		}
		if onLink != nil {
			onLink(href)
		}
	})
	if l.hoverTag == "" {
		return
	}
	hover := l.hoverTag
	tw.TagBind(l.tag, "<Enter>", func() {
		for _, r := range tw.TagRanges(l.tag) {
			_ = tw.TagAdd(hover, r.Start, r.End)
		}
	})
	tw.TagBind(l.tag, "<Leave>", func() {
		_ = tw.TagRemove(hover, "1.0", "end")
	})
}

// loadImage resolves an image source once per document.
func (cfg *config) loadImage(src string) widget.WidgetImage {
	if src == "" {
		return nil
	}
	if img, ok := cfg.images[src]; ok {
		return img
	}
	var img widget.WidgetImage
	var err error
	if cfg.resolver != nil {
		img, err = cfg.resolver(src)
	} else {
		img, err = defaultLoad(src, cfg.baseDir)
	}
	if err != nil {
		img = nil
	}
	cfg.images[src] = img
	return img
}

var errRemote = errors.New("richtext: remote images are not loaded")

// defaultLoad reads data: URIs and local files.
func defaultLoad(src, baseDir string) (widget.WidgetImage, error) {
	if data, ok := strings.CutPrefix(src, "data:"); ok {
		meta, payload, found := strings.Cut(data, ",")
		if !found {
			return nil, errors.New("richtext: bad data URI")
		}
		var raw []byte
		if strings.HasSuffix(meta, ";base64") {
			b, err := base64.StdEncoding.DecodeString(payload)
			if err != nil {
				return nil, err
			}
			raw = b
		} else {
			s, err := url.PathUnescape(payload)
			if err != nil {
				return nil, err
			}
			raw = []byte(s)
		}
		if strings.HasPrefix(meta, "image/svg") {
			return image.NewPhotoFromSVG("richtext-svg", string(raw), 1)
		}
		return image.NewPhotoFromReader("richtext-data", bytes.NewReader(raw))
	}
	if strings.Contains(src, "://") {
		if p, ok := strings.CutPrefix(src, "file://"); ok {
			return image.NewPhotoFromFile(filepath.Base(p), p)
		}
		return nil, errRemote
	}
	path := src
	if !filepath.IsAbs(path) && baseDir != "" {
		path = filepath.Join(baseDir, path)
	}
	if strings.EqualFold(filepath.Ext(path), ".svg") {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return image.NewPhotoFromSVG(filepath.Base(path), string(data), 1)
	}
	return image.NewPhotoFromFile(filepath.Base(path), path)
}

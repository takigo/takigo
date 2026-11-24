package text

// Embedded images: the image create/configure/names commands of
// tk/generic/tkTextImage.c.

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
)

// ImageAlign is an embedded image's -align: where it sits vertically in
// its display line.
type ImageAlign int

const (
	// AlignCenter centres the image in the line (Tk's default).
	AlignCenter ImageAlign = iota
	// AlignTop puts the image's top at the top of the line.
	AlignTop
	// AlignBottom puts the image's bottom at the bottom of the line.
	AlignBottom
	// AlignBaseline puts the image's bottom on the text baseline.
	AlignBaseline
)

// embImage is one embedded image segment (TkTextEmbImage).
type embImage struct {
	img        widget.WidgetImage
	align      ImageAlign
	padX, padY int
	name       string // -name; the image's own name when empty
	float      ImageFloat
}

// instanceName is the name an "image name" index finds the image by.
func (ei *embImage) instanceName() string {
	if ei.name != "" {
		return ei.name
	}
	if n, ok := ei.img.(interface{ Name() string }); ok {
		return n.Name()
	}
	return ""
}

// width is the chunk width: the image plus -padx on each side.
func (ei *embImage) width() int { return ei.img.Width() + 2*ei.padX }

// height is the image plus -pady above and below.
func (ei *embImage) height() int { return ei.img.Height() + 2*ei.padY }

// ImageOption configures an embedded image.
type ImageOption func(*embImage)

// ImageAlignOpt sets where the image sits in its line (-align).
func ImageAlignOpt(a ImageAlign) ImageOption {
	return func(ei *embImage) { ei.align = a }
}

// ImagePadX sets the extra space left and right of the image (-padx).
func ImagePadX[L screenunit.Length](n L) ImageOption {
	return func(ei *embImage) { ei.padX = max(screenunit.ToPixels(n), 0) }
}

// ImagePadY sets the extra space above and below the image (-pady).
func ImagePadY[L screenunit.Length](n L) ImageOption {
	return func(ei *embImage) { ei.padY = max(screenunit.ToPixels(n), 0) }
}

// ImageName sets the name the image is known by in indices ("name") and
// ImageNames (-name).
func ImageName(name string) ImageOption {
	return func(ei *embImage) { ei.name = name }
}

// ImageSource replaces the displayed image (-image).
func ImageSource(img widget.WidgetImage) ImageOption {
	return func(ei *embImage) {
		if img != nil {
			ei.img = img
		}
	}
}

// ImageCreate embeds an image at the given text index. As in Tk it takes
// one index position: text flows around it, it moves with edits before it,
// and deleting its position removes it. Every peer shows it.
//
// As EmbImageConfigure does, a -name already in use gets a "#n" suffix; the
// name actually given is returned by ImageNames.
func (t *TextWidget) ImageCreate[I IndexSpec](indexStr I, img widget.WidgetImage, opts ...ImageOption) error {
	idx, ok := indexOf(t, indexStr)
	if !ok {
		return ErrBadIndex
	}
	if img == nil {
		return nil
	}
	ei := &embImage{img: img}
	for _, opt := range opts {
		opt(ei)
	}
	d := t.doc
	ei.name = d.uniqueImageName(ei)
	d.Insert(idx, string(runeEmbeddedWindow))
	// A mark just before the placeholder, with right gravity so that text
	// inserted at the image's position moves it along.
	d.imageSeq++
	mark := fmt.Sprintf("_ei%d", d.imageSeq)
	d.MarkSet(mark, idx)
	d.addObject(mark)
	if d.images == nil {
		d.images = map[string]*embImage{}
	}
	d.images[mark] = ei
	t.Display()
	return nil
}

// uniqueImageName ports the -name handling of EmbImageConfigure: the base
// name, or name#n with the lowest unused n when the base is taken.
func (d *Document) uniqueImageName(ei *embImage) string {
	base := ei.instanceName()
	if base == "" {
		return ""
	}
	taken := map[string]bool{}
	d.eachImage(func(_ Index, o *embImage) { taken[o.instanceName()] = true })
	if !taken[base] {
		return base
	}
	for n := 1; ; n++ {
		if name := base + "#" + strconv.Itoa(n); !taken[name] {
			return name
		}
	}
}

// ImageConfigure changes the options of the embedded image at index
// (image configure).
func (t *TextWidget) ImageConfigure[I IndexSpec](index I, opts ...ImageOption) error {
	idx, ok := indexOf(t, index)
	if !ok {
		return ErrBadIndex
	}
	ei := t.doc.imageAt(idx)
	if ei == nil {
		return fmt.Errorf("no embedded image at index %q: %w", fmt.Sprint(index), ErrBadIndex)
	}
	for _, opt := range opts {
		opt(ei)
	}
	t.doc.layoutGen++
	t.doc.notifyListeners(Change{From: idx.Line, To: idx.Line})
	return nil
}

// ImageAt returns the image embedded at index and its options.
func (t *TextWidget) ImageAt[I IndexSpec](index I) (img widget.WidgetImage, align ImageAlign, padX, padY int, ok bool) {
	idx, found := indexOf(t, index)
	if !found {
		return nil, 0, 0, 0, false
	}
	ei := t.doc.imageAt(idx)
	if ei == nil {
		return nil, 0, 0, 0, false
	}
	return ei.img, ei.align, ei.padX, ei.padY, true
}

// ImageNames returns the names of the embedded images in document order
// (image names).
func (t *TextWidget) ImageNames() []string {
	type named struct {
		pos  Index
		name string
	}
	var all []named
	t.doc.eachImage(func(pos Index, ei *embImage) {
		all = append(all, named{pos, ei.instanceName()})
	})
	slices.SortFunc(all, func(a, b named) int { return Compare(a.pos, b.pos) })
	names := make([]string, 0, len(all))
	for _, n := range all {
		if n.name != "" {
			names = append(names, n.name)
		}
	}
	return names
}

// ParseImageAlign reads Tk's -align values: top, center, bottom, baseline.
func ParseImageAlign(s string) (ImageAlign, error) {
	switch strings.ToLower(s) {
	case "top":
		return AlignTop, nil
	case "center", "middle":
		return AlignCenter, nil
	case "bottom":
		return AlignBottom, nil
	case "baseline":
		return AlignBaseline, nil
	}
	return AlignCenter, fmt.Errorf("bad align %q: must be baseline, bottom, center, or top", s)
}

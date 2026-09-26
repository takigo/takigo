// Package nanosvg ports the parts of nanosvg.h / nanosvgrast.h (vendored in
// tk/generic) that Tk 9 uses to render its built-in SVG images: check and
// radio indicators, sliders, toggle switches. Only the SVG subset those images
// use is parsed (path with M/L/H/V/Z, rect with rx/ry, circle, ellipse,
// linearGradient in user space, solid fills and strokes). Arithmetic is kept
// in float32 so that the rasterized pixels match Tk's exactly.
package nanosvg

import (
	"encoding/xml"
	"math"
	"strconv"
	"strings"
)

const kappa90 = float32(0.5522847493)

type paintType int8

const (
	paintNone paintType = iota
	paintColor
	paintLinearGradient
)

// LineCap and LineJoin follow NSVGlineCap / NSVGlineJoin.
const (
	capButt = iota
	capRound
	capSquare
)

const (
	joinMiter = iota
	joinRound
	joinBevel
)

type gradientStop struct {
	color  uint32
	offset float32
}

type gradient struct {
	xform [6]float32
	stops []gradientStop
}

type paint struct {
	typ      paintType
	color    uint32
	gradient *gradient
}

type path struct {
	pts    []float32 // x0,y0, then cubic segments (3 points each)
	closed bool
}

type shape struct {
	fill, stroke   paint
	opacity        float32
	strokeWidth    float32
	strokeLineJoin int
	strokeLineCap  int
	miterLimit     float32
	evenOdd        bool
	paths          []*path
}

// Image is a parsed SVG document.
type Image struct {
	Width, Height float32
	shapes        []*shape
}

type gradientData struct {
	id             string
	x1, y1, x2, y2 float32
	stops          []gradientStop
}

type attrs struct {
	fill, stroke            string
	fillOpacity, strokeOp   float32
	opacity                 float32
	strokeWidth             float32
	lineJoin, lineCap       int
	miterLimit              float32
	evenOdd                 bool
	stopColor               uint32
	stopOpacity, stopOffset float32
}

func defaultAttrs() attrs {
	return attrs{fill: "#000000", stroke: "none", fillOpacity: 1, strokeOp: 1, opacity: 1,
		strokeWidth: 1, lineJoin: joinMiter, lineCap: capButt, miterLimit: 4, stopOpacity: 1}
}

func atof(s string) float32 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return float32(v)
}

// parseColor handles the "#rrggbb" and "#rgb" forms nsvg__parseColorHex does;
// colours are packed as r | g<<8 | b<<16 like NSVG_RGB.
func parseColor(s string) uint32 {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "#") {
		h := s[1:]
		if len(h) >= 6 {
			if v, err := strconv.ParseUint(h[:6], 16, 32); err == nil {
				return uint32(v>>16&0xff) | uint32(v>>8&0xff)<<8 | uint32(v&0xff)<<16
			}
		}
		if len(h) >= 3 {
			if v, err := strconv.ParseUint(h[:3], 16, 32); err == nil {
				r, g, b := uint32(v>>8&0xf)*17, uint32(v>>4&0xf)*17, uint32(v&0xf)*17
				return r | g<<8 | b<<16
			}
		}
	}
	switch s {
	case "black":
		return 0
	case "white":
		return 0xffffff
	}
	return 128 | 128<<8 | 128<<16
}

func (a *attrs) set(name, value string) bool {
	switch name {
	case "fill":
		a.fill = value
	case "stroke":
		a.stroke = value
	case "fill-opacity":
		a.fillOpacity = clampf(atof(value), 0, 1)
	case "stroke-opacity":
		a.strokeOp = clampf(atof(value), 0, 1)
	case "opacity":
		a.opacity = clampf(atof(value), 0, 1)
	case "stroke-width":
		a.strokeWidth = atof(value)
	case "stroke-linecap":
		switch value {
		case "round":
			a.lineCap = capRound
		case "square":
			a.lineCap = capSquare
		default:
			a.lineCap = capButt
		}
	case "stroke-linejoin":
		switch value {
		case "round":
			a.lineJoin = joinRound
		case "bevel":
			a.lineJoin = joinBevel
		default:
			a.lineJoin = joinMiter
		}
	case "stroke-miterlimit":
		a.miterLimit = atof(value)
	case "fill-rule":
		a.evenOdd = value == "evenodd"
	case "stop-color":
		a.stopColor = parseColor(value)
	case "stop-opacity":
		a.stopOpacity = clampf(atof(value), 0, 1)
	case "offset":
		v := strings.TrimSpace(value)
		if strings.HasSuffix(v, "%") {
			a.stopOffset = atof(strings.TrimSuffix(v, "%")) / 100
		} else {
			a.stopOffset = atof(v)
		}
	default:
		return false
	}
	return true
}

type parser struct {
	img       *Image
	gradients map[string]*gradientData
	pending   []*pendingShape
	pts       []float32
	paths     []*path
}

type pendingShape struct {
	s                  *shape
	fillRef, strokeRef string
}

// Parse parses an SVG document in the subset Tk's built-in images use.
func Parse(src string) (*Image, error) {
	p := &parser{img: &Image{}, gradients: map[string]*gradientData{}}
	dec := xml.NewDecoder(strings.NewReader(src))
	stack := []attrs{defaultAttrs()}
	var curGrad *gradientData
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			a := stack[len(stack)-1]
			get := func(n string) (string, bool) {
				for _, at := range t.Attr {
					if at.Name.Local == n {
						return at.Value, true
					}
				}
				return "", false
			}
			switch t.Name.Local {
			case "svg":
				if v, ok := get("width"); ok {
					p.img.Width = atof(strings.TrimSuffix(v, "px"))
				}
				if v, ok := get("height"); ok {
					p.img.Height = atof(strings.TrimSuffix(v, "px"))
				}
			case "linearGradient":
				g := &gradientData{x2: 1}
				g.id, _ = get("id")
				for _, at := range t.Attr {
					switch at.Name.Local {
					case "x1":
						g.x1 = atof(at.Value)
					case "y1":
						g.y1 = atof(at.Value)
					case "x2":
						g.x2 = atof(at.Value)
					case "y2":
						g.y2 = atof(at.Value)
					}
				}
				p.gradients[g.id] = g
				curGrad = g
			case "stop":
				sa := a
				sa.stopOffset, sa.stopColor, sa.stopOpacity = 0, 0, 1
				for _, at := range t.Attr {
					sa.set(at.Name.Local, at.Value)
				}
				if curGrad != nil {
					st := gradientStop{color: sa.stopColor | uint32(sa.stopOpacity*255)<<24, offset: sa.stopOffset}
					idx := len(curGrad.stops)
					for i, o := range curGrad.stops {
						if st.offset < o.offset {
							idx = i
							break
						}
					}
					curGrad.stops = append(curGrad.stops, gradientStop{})
					copy(curGrad.stops[idx+1:], curGrad.stops[idx:])
					curGrad.stops[idx] = st
				}
			default:
				for _, at := range t.Attr {
					a.set(at.Name.Local, at.Value)
				}
				switch t.Name.Local {
				case "path":
					d, _ := get("d")
					p.parsePath(d, a)
				case "rect":
					p.parseRect(t, a)
				case "circle":
					cx, cy, r := attrF(t, "cx"), attrF(t, "cy"), float32(math.Abs(float64(attrF(t, "r"))))
					if r > 0 {
						p.ellipse(cx, cy, r, r, a)
					}
				case "ellipse":
					cx, cy := attrF(t, "cx"), attrF(t, "cy")
					rx, ry := float32(math.Abs(float64(attrF(t, "rx")))), float32(math.Abs(float64(attrF(t, "ry"))))
					if rx > 0 && ry > 0 {
						p.ellipse(cx, cy, rx, ry, a)
					}
				}
			}
			stack = append(stack, a)
		case xml.EndElement:
			if t.Name.Local == "linearGradient" {
				curGrad = nil
			}
			stack = stack[:len(stack)-1]
		}
	}
	p.createGradients()
	return p.img, nil
}

func attrF(t xml.StartElement, n string) float32 {
	for _, at := range t.Attr {
		if at.Name.Local == n {
			return atof(at.Value)
		}
	}
	return 0
}

func (p *parser) moveTo(x, y float32) {
	if len(p.pts) > 0 {
		p.pts[len(p.pts)-2], p.pts[len(p.pts)-1] = x, y
	} else {
		p.pts = append(p.pts, x, y)
	}
}

func (p *parser) lineTo(x, y float32) {
	if len(p.pts) > 0 {
		px, py := p.pts[len(p.pts)-2], p.pts[len(p.pts)-1]
		dx, dy := x-px, y-py
		p.pts = append(p.pts, px+dx/3, py+dy/3, x-dx/3, y-dy/3, x, y)
	}
}

func (p *parser) cubicBezTo(cx1, cy1, cx2, cy2, x, y float32) {
	if len(p.pts) > 0 {
		p.pts = append(p.pts, cx1, cy1, cx2, cy2, x, y)
	}
}

func (p *parser) addPath(closed bool) {
	if len(p.pts)/2 < 4 {
		return
	}
	if closed {
		p.lineTo(p.pts[0], p.pts[1])
	}
	if (len(p.pts)/2)%3 != 1 {
		return
	}
	p.paths = append(p.paths, &path{pts: append([]float32(nil), p.pts...), closed: closed})
}

func (p *parser) addShape(a attrs) {
	if len(p.paths) == 0 {
		return
	}
	s := &shape{
		opacity: a.opacity, strokeWidth: a.strokeWidth,
		strokeLineJoin: a.lineJoin, strokeLineCap: a.lineCap, miterLimit: a.miterLimit,
		evenOdd: a.evenOdd, paths: p.paths,
	}
	p.paths = nil
	ps := &pendingShape{s: s}
	setPaint := func(spec string, op float32, dst *paint, ref *string) {
		spec = strings.TrimSpace(spec)
		switch {
		case spec == "none":
			dst.typ = paintNone
		case strings.HasPrefix(spec, "url("):
			id := strings.TrimSuffix(strings.TrimPrefix(spec, "url("), ")")
			*ref = strings.TrimPrefix(strings.TrimSpace(id), "#")
		default:
			dst.typ = paintColor
			dst.color = parseColor(spec) | uint32(op*255)<<24
		}
	}
	setPaint(a.fill, a.fillOpacity, &s.fill, &ps.fillRef)
	setPaint(a.stroke, a.strokeOp, &s.stroke, &ps.strokeRef)
	p.pending = append(p.pending, ps)
	p.img.shapes = append(p.img.shapes, s)
}

// createGradients ports nsvg__createGradient for userSpaceOnUse linear
// gradients under an identity transform, followed by the inversion that
// nsvg__scaleToViewbox applies.
func (p *parser) createGradients() {
	mk := func(id string, dst *paint) {
		g := p.gradients[id]
		if g == nil || len(g.stops) == 0 {
			dst.typ = paintNone
			return
		}
		dx, dy := g.x2-g.x1, g.y2-g.y1
		t := [6]float32{dy, -dx, dx, dy, g.x1, g.y1}
		dst.typ = paintLinearGradient
		dst.gradient = &gradient{xform: xformInverse(t), stops: g.stops}
	}
	for _, ps := range p.pending {
		if ps.fillRef != "" {
			mk(ps.fillRef, &ps.s.fill)
		}
		if ps.strokeRef != "" {
			mk(ps.strokeRef, &ps.s.stroke)
		}
	}
}

func xformInverse(t [6]float32) [6]float32 {
	det := float64(t[0])*float64(t[3]) - float64(t[2])*float64(t[1])
	if det > -1e-6 && det < 1e-6 {
		return [6]float32{1, 0, 0, 1, 0, 0}
	}
	invdet := 1.0 / det
	return [6]float32{
		float32(float64(t[3]) * invdet),
		float32(float64(-t[1]) * invdet),
		float32(float64(-t[2]) * invdet),
		float32(float64(t[0]) * invdet),
		float32((float64(t[2])*float64(t[5]) - float64(t[3])*float64(t[4])) * invdet),
		float32((float64(t[1])*float64(t[4]) - float64(t[0])*float64(t[5])) * invdet),
	}
}

func (p *parser) parseRect(t xml.StartElement, a attrs) {
	x, y, w, h := attrF(t, "x"), attrF(t, "y"), attrF(t, "width"), attrF(t, "height")
	rx, ry := float32(-1), float32(-1)
	for _, at := range t.Attr {
		switch at.Name.Local {
		case "rx":
			rx = float32(math.Abs(float64(atof(at.Value))))
		case "ry":
			ry = float32(math.Abs(float64(atof(at.Value))))
		}
	}
	if rx < 0 && ry > 0 {
		rx = ry
	}
	if ry < 0 && rx > 0 {
		ry = rx
	}
	rx, ry = max(rx, 0), max(ry, 0)
	rx, ry = min(rx, w/2), min(ry, h/2)
	if w == 0 || h == 0 {
		return
	}
	p.pts = p.pts[:0]
	if rx < 0.00001 || ry < 0.0001 {
		p.moveTo(x, y)
		p.lineTo(x+w, y)
		p.lineTo(x+w, y+h)
		p.lineTo(x, y+h)
	} else {
		k := 1 - kappa90
		p.moveTo(x+rx, y)
		p.lineTo(x+w-rx, y)
		p.cubicBezTo(x+w-rx*k, y, x+w, y+ry*k, x+w, y+ry)
		p.lineTo(x+w, y+h-ry)
		p.cubicBezTo(x+w, y+h-ry*k, x+w-rx*k, y+h, x+w-rx, y+h)
		p.lineTo(x+rx, y+h)
		p.cubicBezTo(x+rx*k, y+h, x, y+h-ry*k, x, y+h-ry)
		p.lineTo(x, y+ry)
		p.cubicBezTo(x, y+ry*k, x+rx*k, y, x+rx, y)
	}
	p.addPath(true)
	p.addShape(a)
}

func (p *parser) ellipse(cx, cy, rx, ry float32, a attrs) {
	p.pts = p.pts[:0]
	p.moveTo(cx+rx, cy)
	p.cubicBezTo(cx+rx, cy+ry*kappa90, cx+rx*kappa90, cy+ry, cx, cy+ry)
	p.cubicBezTo(cx-rx*kappa90, cy+ry, cx-rx, cy+ry*kappa90, cx-rx, cy)
	p.cubicBezTo(cx-rx, cy-ry*kappa90, cx-rx*kappa90, cy-ry, cx, cy-ry)
	p.cubicBezTo(cx+rx*kappa90, cy-ry, cx+rx, cy-ry*kappa90, cx+rx, cy)
	p.addPath(true)
	p.addShape(a)
}

// pathItems splits path data like nsvg__getNextPathItem: numbers and single
// command letters.
func pathItems(d string) []string {
	var items []string
	i := 0
	for i < len(d) {
		c := d[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ',' {
			i++
			continue
		}
		if c == '-' || c == '+' || c == '.' || (c >= '0' && c <= '9') {
			j := i
			if d[j] == '-' || d[j] == '+' {
				j++
			}
			for j < len(d) && d[j] >= '0' && d[j] <= '9' {
				j++
			}
			if j < len(d) && d[j] == '.' {
				j++
				for j < len(d) && d[j] >= '0' && d[j] <= '9' {
					j++
				}
			}
			if j < len(d) && (d[j] == 'e' || d[j] == 'E') && j+1 < len(d) && d[j+1] != 'm' && d[j+1] != 'x' {
				j++
				if j < len(d) && (d[j] == '-' || d[j] == '+') {
					j++
				}
				for j < len(d) && d[j] >= '0' && d[j] <= '9' {
					j++
				}
			}
			items = append(items, d[i:j])
			i = j
			continue
		}
		items = append(items, d[i:i+1])
		i++
	}
	return items
}

func argsPer(cmd byte) int {
	switch cmd {
	case 'v', 'V', 'h', 'H':
		return 1
	case 'm', 'M', 'l', 'L', 't', 'T':
		return 2
	case 'q', 'Q', 's', 'S':
		return 4
	case 'c', 'C':
		return 6
	case 'a', 'A':
		return 7
	case 'z', 'Z':
		return 0
	}
	return -1
}

// parsePath ports nsvg__parsePath for the M/L/H/V/C/Z commands.
func (p *parser) parsePath(d string, a attrs) {
	var cmd byte
	var args [10]float32
	nargs, rargs := 0, 0
	var cpx, cpy float32
	initPoint, closed := false, false
	p.pts = p.pts[:0]
	for _, item := range pathItems(d) {
		c := item[0]
		isCoord := c == '-' || c == '+' || c == '.' || (c >= '0' && c <= '9')
		if cmd != 0 && isCoord {
			if nargs < 10 {
				args[nargs] = atof(item)
				nargs++
			}
			if nargs >= rargs {
				rel := cmd >= 'a'
				switch cmd {
				case 'm', 'M':
					if rel {
						cpx += args[0]
						cpy += args[1]
					} else {
						cpx, cpy = args[0], args[1]
					}
					p.moveTo(cpx, cpy)
					if rel {
						cmd = 'l'
					} else {
						cmd = 'L'
					}
					rargs = argsPer(cmd)
					initPoint = true
				case 'l', 'L':
					if rel {
						cpx += args[0]
						cpy += args[1]
					} else {
						cpx, cpy = args[0], args[1]
					}
					p.lineTo(cpx, cpy)
				case 'h', 'H':
					if rel {
						cpx += args[0]
					} else {
						cpx = args[0]
					}
					p.lineTo(cpx, cpy)
				case 'v', 'V':
					if rel {
						cpy += args[0]
					} else {
						cpy = args[0]
					}
					p.lineTo(cpx, cpy)
				case 'c', 'C':
					x1, y1, x2, y2, x, y := args[0], args[1], args[2], args[3], args[4], args[5]
					if rel {
						x1, y1, x2, y2, x, y = x1+cpx, y1+cpy, x2+cpx, y2+cpy, x+cpx, y+cpy
					}
					p.cubicBezTo(x1, y1, x2, y2, x, y)
					cpx, cpy = x, y
				default:
					if nargs >= 2 {
						cpx, cpy = args[nargs-2], args[nargs-1]
					}
				}
				nargs = 0
			}
			continue
		}
		cmd = c
		if cmd == 'M' || cmd == 'm' {
			if len(p.pts) > 0 {
				p.addPath(closed)
			}
			p.pts = p.pts[:0]
			closed = false
			nargs = 0
		} else if !initPoint {
			cmd = 0
		}
		if cmd == 'Z' || cmd == 'z' {
			closed = true
			if len(p.pts) > 0 {
				cpx, cpy = p.pts[0], p.pts[1]
				p.addPath(closed)
			}
			p.pts = p.pts[:0]
			p.moveTo(cpx, cpy)
			closed = false
			nargs = 0
		}
		rargs = argsPer(cmd)
		if rargs == -1 {
			cmd, rargs = 0, 0
		}
	}
	if len(p.pts) > 0 {
		p.addPath(closed)
	}
	p.addShape(a)
}

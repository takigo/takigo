// Package nanosvg ports the parts of nanosvg.h / nanosvgrast.h (vendored in
// tk/generic) that Tk 9 uses to render its built-in SVG images: check and
// radio indicators, sliders, toggle switches, dialog icons. Only the SVG
// subset those images use is parsed (path with every command, rect with
// rx/ry, circle, ellipse, linearGradient in user space, solid fills and
// strokes). Arithmetic is kept
// in float32 so that the rasterized pixels match Tk's exactly.
package nanosvg

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
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
	xform                   [6]float32 // user space → document, as nsvg's attr.xform
}

func defaultAttrs() attrs {
	return attrs{fill: "#000000", stroke: "none", fillOpacity: 1, strokeOp: 1, opacity: 1,
		strokeWidth: 1, lineJoin: joinMiter, lineCap: capButt, miterLimit: 4, stopOpacity: 1,
		xform: identity}
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

func (a *attrs) set(name, value string) {
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
		if before, ok := strings.CutSuffix(v, "%"); ok {
			a.stopOffset = atof(before) / 100
		} else {
			a.stopOffset = atof(v)
		}
	}
}

type parser struct {
	img       *Image
	gradients map[string]*gradientData
	pending   []*pendingShape
	pts       []float32
	paths     []*path
	xform     [6]float32 // transform of the element whose path is being built
}

type pendingShape struct {
	s                  *shape
	fillRef, strokeRef string
}

// Parse parses an SVG document in the subset Tk's built-in images use.
// A malformed document is reported as an error; the image parsed so far
// is returned with it, so that callers that ignore the error still have
// something to work with.
func Parse(src string) (*Image, error) {
	p := &parser{img: &Image{}, gradients: map[string]*gradientData{}}
	dec := xml.NewDecoder(strings.NewReader(src))
	stack := []attrs{defaultAttrs()}
	var curGrad *gradientData
	root := false
	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return p.img, fmt.Errorf("nanosvg: parse svg: %w", err)
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
				root = true
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
				// Presentation attributes first, then style declarations,
				// which take precedence (nsvg__parseAttribs).
				for _, at := range t.Attr {
					switch at.Name.Local {
					case "style":
					case "transform":
						a.xform = xformMultiply(a.xform, parseTransform(at.Value))
					default:
						a.set(at.Name.Local, at.Value)
					}
				}
				if v, ok := get("style"); ok {
					for decl := range strings.SplitSeq(v, ";") {
						if name, value, ok := strings.Cut(decl, ":"); ok {
							a.set(strings.TrimSpace(name), strings.TrimSpace(value))
						}
					}
				}
				p.xform = a.xform
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
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	if !root {
		return p.img, errors.New("nanosvg: parse svg: no <svg> element")
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
	pts := append([]float32(nil), p.pts...)
	if p.xform != identity {
		t := p.xform
		for i := 0; i < len(pts); i += 2 {
			x, y := pts[i], pts[i+1]
			pts[i], pts[i+1] = x*t[0]+y*t[2]+t[4], x*t[1]+y*t[3]+t[5]
		}
	}
	p.paths = append(p.paths, &path{pts: pts, closed: closed})
}

func (p *parser) addShape(a attrs) {
	if len(p.paths) == 0 {
		return
	}
	s := &shape{
		opacity: a.opacity, strokeWidth: a.strokeWidth * averageScale(a.xform),
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

// nextPathItem ports nsvg__getNextPathItem: it returns the number or single
// command letter at d[i:] and the index after it, or "" at the end. With
// arcFlag (the large-arc and sweep arguments of A) a lone 0 or 1 is one
// item, so compact flags such as "a2 2 0 011 1" parse.
func nextPathItem(d string, i int, arcFlag bool) (string, int) {
	for i < len(d) && (d[i] == ' ' || d[i] == '\t' || d[i] == '\n' || d[i] == '\r' || d[i] == '\f' || d[i] == '\v' || d[i] == ',') {
		i++
	}
	if i >= len(d) {
		return "", i
	}
	if arcFlag {
		if d[i] == '0' || d[i] == '1' {
			return d[i : i+1], i + 1
		}
	}
	c := d[i]
	if c != '-' && c != '+' && c != '.' && (c < '0' || c > '9') {
		return d[i : i+1], i + 1
	}
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
	return d[i:j], j
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

// parsePath ports nsvg__parsePath. (cpx2, cpy2) is the last control point,
// which the S and T shorthands reflect.
func (p *parser) parsePath(d string, a attrs) {
	var cmd byte
	var args [10]float32
	nargs, rargs := 0, 0
	var cpx, cpy, cpx2, cpy2 float32
	initPoint, closed := false, false
	p.pts = p.pts[:0]
	for i := 0; ; {
		var item string
		item, i = nextPathItem(d, i, (cmd == 'A' || cmd == 'a') && (nargs == 3 || nargs == 4))
		if item == "" {
			break
		}
		c := item[0]
		isCoord := c == '-' || c == '+' || c == '.' || (c >= '0' && c <= '9')
		if cmd != 0 && isCoord {
			if nargs < 10 {
				args[nargs] = atof(item)
				nargs++
			}
			if nargs >= rargs {
				rel := cmd >= 'a'
				var ox, oy float32 // origin of relative coordinates
				if rel {
					ox, oy = cpx, cpy
				}
				switch cmd {
				case 'm', 'M':
					cpx, cpy = ox+args[0], oy+args[1]
					p.moveTo(cpx, cpy)
					if rel {
						cmd = 'l'
					} else {
						cmd = 'L'
					}
					rargs = argsPer(cmd)
					cpx2, cpy2 = cpx, cpy
					initPoint = true
				case 'l', 'L':
					cpx, cpy = ox+args[0], oy+args[1]
					p.lineTo(cpx, cpy)
					cpx2, cpy2 = cpx, cpy
				case 'h', 'H':
					cpx = ox + args[0]
					p.lineTo(cpx, cpy)
					cpx2, cpy2 = cpx, cpy
				case 'v', 'V':
					cpy = oy + args[0]
					p.lineTo(cpx, cpy)
					cpx2, cpy2 = cpx, cpy
				case 'c', 'C':
					x1, y1, x2, y2 := ox+args[0], oy+args[1], ox+args[2], oy+args[3]
					x, y := ox+args[4], oy+args[5]
					p.cubicBezTo(x1, y1, x2, y2, x, y)
					cpx2, cpy2, cpx, cpy = x2, y2, x, y
				case 's', 'S':
					x2, y2, x, y := ox+args[0], oy+args[1], ox+args[2], oy+args[3]
					p.cubicBezTo(2*cpx-cpx2, 2*cpy-cpy2, x2, y2, x, y)
					cpx2, cpy2, cpx, cpy = x2, y2, x, y
				case 'q', 'Q':
					qx, qy, x, y := ox+args[0], oy+args[1], ox+args[2], oy+args[3]
					p.quadBezTo(cpx, cpy, qx, qy, x, y)
					cpx2, cpy2, cpx, cpy = qx, qy, x, y
				case 't', 'T':
					qx, qy, x, y := 2*cpx-cpx2, 2*cpy-cpy2, ox+args[0], oy+args[1]
					p.quadBezTo(cpx, cpy, qx, qy, x, y)
					cpx2, cpy2, cpx, cpy = qx, qy, x, y
				case 'a', 'A':
					x, y := ox+args[5], oy+args[6]
					p.arcTo(cpx, cpy, args[0], args[1], args[2], args[3], args[4], x, y)
					cpx, cpy = x, y
					cpx2, cpy2 = cpx, cpy
				default:
					if nargs >= 2 {
						cpx, cpy = args[nargs-2], args[nargs-1]
						cpx2, cpy2 = cpx, cpy
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
				cpx2, cpy2 = cpx, cpy
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

// quadBezTo adds the quadratic curve from (x1, y1) through control point
// (cx, cy) to (x2, y2) as the cubic nsvg__pathQuadBezTo converts it to.
func (p *parser) quadBezTo(x1, y1, cx, cy, x2, y2 float32) {
	p.cubicBezTo(x1+2.0/3.0*(cx-x1), y1+2.0/3.0*(cy-y1),
		x2+2.0/3.0*(cx-x2), y2+2.0/3.0*(cy-y2), x2, y2)
}

func vecang(ux, uy, vx, vy float32) float32 {
	r := (ux*vx + uy*vy) / (sqrtf(ux*ux+uy*uy) * sqrtf(vx*vx+vy*vy))
	r = clampf(r, -1, 1)
	s := float32(1)
	if ux*vy < uy*vx {
		s = -1
	}
	return s * float32(math.Acos(float64(r)))
}

// arcTo ports nsvg__pathArcTo: the elliptical arc from (x1, y1) to (x2, y2)
// as cubic segments of at most 90 degrees.
func (p *parser) arcTo(x1, y1, rx, ry, rotDeg, large, sweep, x2, y2 float32) {
	const pi = float32(math.Pi)
	rx, ry = float32(math.Abs(float64(rx))), float32(math.Abs(float64(ry)))
	rotx := rotDeg / 180 * pi
	fa := math.Abs(float64(large)) > 1e-6
	fs := math.Abs(float64(sweep)) > 1e-6

	dx, dy := x1-x2, y1-y2
	d := sqrtf(dx*dx + dy*dy)
	if d < 1e-6 || rx < 1e-6 || ry < 1e-6 {
		p.lineTo(x2, y2)
		return
	}
	sinrx, cosrx := sinf(rotx), cosf(rotx)

	x1p := cosrx*dx/2 + sinrx*dy/2
	y1p := -sinrx*dx/2 + cosrx*dy/2
	d = x1p*x1p/(rx*rx) + y1p*y1p/(ry*ry)
	if d > 1 {
		d = sqrtf(d)
		rx *= d
		ry *= d
	}
	var s float32
	sa := rx*rx*ry*ry - rx*rx*y1p*y1p - ry*ry*x1p*x1p
	sb := rx*rx*y1p*y1p + ry*ry*x1p*x1p
	if sa < 0 {
		sa = 0
	}
	if sb > 0 {
		s = sqrtf(sa / sb)
	}
	if fa == fs {
		s = -s
	}
	cxp := s * rx * y1p / ry
	cyp := s * -ry * x1p / rx

	cx := (x1+x2)/2 + cosrx*cxp - sinrx*cyp
	cy := (y1+y2)/2 + sinrx*cxp + cosrx*cyp

	ux, uy := (x1p-cxp)/rx, (y1p-cyp)/ry
	vx, vy := (-x1p-cxp)/rx, (-y1p-cyp)/ry
	a1 := vecang(1, 0, ux, uy)
	da := vecang(ux, uy, vx, vy)
	if !fs && da > 0 {
		da -= 2 * pi
	} else if fs && da < 0 {
		da += 2 * pi
	}

	t := [6]float32{cosrx, sinrx, -sinrx, cosrx, cx, cy}
	ndivs := int(float32(math.Abs(float64(da)))/(pi*0.5) + 1)
	hda := (da / float32(ndivs)) / 2
	if hda < 1e-3 && hda > -1e-3 {
		hda *= 0.5
	} else {
		hda = (1 - cosf(hda)) / sinf(hda)
	}
	kappa := float32(math.Abs(float64(4.0 / 3.0 * hda)))
	if da < 0 {
		kappa = -kappa
	}

	var px, py, ptanx, ptany float32
	for i := 0; i <= ndivs; i++ {
		ang := a1 + da*(float32(i)/float32(ndivs))
		dx, dy := cosf(ang), sinf(ang)
		x := dx*rx*t[0] + dy*ry*t[2] + t[4]
		y := dx*rx*t[1] + dy*ry*t[3] + t[5]
		vx, vy := -dy*rx*kappa, dx*ry*kappa
		tanx := vx*t[0] + vy*t[2]
		tany := vx*t[1] + vy*t[3]
		if i > 0 {
			p.cubicBezTo(px+ptanx, py+ptany, x-tanx, y-tany, x, y)
		}
		px, py, ptanx, ptany = x, y, tanx, tany
	}
}

// identity is the identity affine transform [a b c d e f], mapping
// (x, y) to (a*x + c*y + e, b*x + d*y + f) as nanosvg's xform arrays do.
var identity = [6]float32{1, 0, 0, 1, 0, 0}

// xformMultiply returns the transform applying t first, then parent.
func xformMultiply(parent, t [6]float32) [6]float32 {
	return [6]float32{
		t[0]*parent[0] + t[1]*parent[2],
		t[0]*parent[1] + t[1]*parent[3],
		t[2]*parent[0] + t[3]*parent[2],
		t[2]*parent[1] + t[3]*parent[3],
		t[4]*parent[0] + t[5]*parent[2] + parent[4],
		t[4]*parent[1] + t[5]*parent[3] + parent[5],
	}
}

// averageScale ports nsvg__getAverageScale, used to scale stroke widths.
func averageScale(t [6]float32) float32 {
	sx := math.Hypot(float64(t[0]), float64(t[2]))
	sy := math.Hypot(float64(t[1]), float64(t[3]))
	return float32((sx + sy) * 0.5)
}

// parseTransform parses an SVG transform list (matrix, translate, scale,
// rotate, skewX, skewY) as nsvg__parseTransform does; unknown functions
// are ignored.
func parseTransform(s string) [6]float32 {
	xf := identity
	for {
		open := strings.IndexByte(s, '(')
		end := strings.IndexByte(s, ')')
		if open < 0 || end < open {
			return xf
		}
		name := strings.TrimLeft(s[:open], " \t\n\r,")
		var args []float32
		for f := range strings.FieldsFuncSeq(s[open+1:end], func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
		}) {
			args = append(args, atof(f))
		}
		arg := func(i int, def float32) float32 {
			if i < len(args) {
				return args[i]
			}
			return def
		}
		t := identity
		switch name {
		case "matrix":
			if len(args) == 6 {
				copy(t[:], args)
			}
		case "translate":
			t[4], t[5] = arg(0, 0), arg(1, 0)
		case "scale":
			t[0] = arg(0, 1)
			t[3] = arg(1, t[0])
		case "rotate":
			a := float64(arg(0, 0)) * math.Pi / 180
			sin, cos := float32(math.Sin(a)), float32(math.Cos(a))
			cx, cy := arg(1, 0), arg(2, 0)
			t = [6]float32{cos, sin, -sin, cos, 0, 0}
			if cx != 0 || cy != 0 {
				t = xformMultiply([6]float32{1, 0, 0, 1, cx, cy}, xformMultiply(t, [6]float32{1, 0, 0, 1, -cx, -cy}))
			}
		case "skewX":
			t[2] = float32(math.Tan(float64(arg(0, 0)) * math.Pi / 180))
		case "skewY":
			t[1] = float32(math.Tan(float64(arg(0, 0)) * math.Pi / 180))
		}
		xf = xformMultiply(xf, t)
		s = s[end+1:]
	}
}

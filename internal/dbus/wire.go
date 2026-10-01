// Package dbus is a minimal D-Bus client: enough to make method calls on
// the session bus (desktop notifications, the settings portal) without a
// third-party module. It implements the wire format of the D-Bus
// specification for the types those calls use; it does not export objects
// or dispatch signals.
package dbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// Variant is a D-Bus variant: a value with its own signature.
type Variant struct {
	Sig   string
	Value any
}

// ObjectPath is a D-Bus object path (signature "o").
type ObjectPath string

// ErrBadMessage is wrapped by errors for data that does not match its
// signature or is truncated.
var ErrBadMessage = errors.New("dbus: malformed message")

// alignment returns the alignment of the type starting at sig[0].
func alignment(c byte) int {
	switch c {
	case 'y', 'g', 'v':
		return 1
	case 'n', 'q':
		return 2
	case 'b', 'i', 'u', 's', 'o', 'a':
		return 4
	default: // x t d ( {
		return 8
	}
}

// next returns the first complete type of sig and the rest.
func next(sig string) (string, string, error) {
	if sig == "" {
		return "", "", fmt.Errorf("%w: empty signature", ErrBadMessage)
	}
	switch sig[0] {
	case 'a':
		elem, rest, err := next(sig[1:])
		if err != nil {
			return "", "", err
		}
		return sig[:1+len(elem)], rest, nil
	case '(', '{':
		closer := byte(')')
		if sig[0] == '{' {
			closer = '}'
		}
		depth := 0
		for i := 0; i < len(sig); i++ {
			switch sig[i] {
			case sig[0]:
				depth++
			case closer:
				depth--
				if depth == 0 {
					return sig[:i+1], sig[i+1:], nil
				}
			}
		}
		return "", "", fmt.Errorf("%w: unbalanced %q", ErrBadMessage, sig)
	}
	return sig[:1], sig[1:], nil
}

type encoder struct {
	buf []byte
}

func (e *encoder) align(n int) {
	for len(e.buf)%n != 0 {
		e.buf = append(e.buf, 0)
	}
}

func (e *encoder) u32(v uint32) {
	e.align(4)
	e.buf = binary.LittleEndian.AppendUint32(e.buf, v)
}

func (e *encoder) str(s string) {
	e.u32(uint32(len(s)))
	e.buf = append(append(e.buf, s...), 0)
}

func (e *encoder) sig(s string) {
	e.buf = append(append(append(e.buf, byte(len(s))), s...), 0)
}

// values encodes vals, one per complete type of sig.
func (e *encoder) values(sig string, vals []any) error {
	for _, v := range vals {
		t, rest, err := next(sig)
		if err != nil {
			return err
		}
		if err := e.value(t, v); err != nil {
			return err
		}
		sig = rest
	}
	if sig != "" {
		return fmt.Errorf("%w: no value for %q", ErrBadMessage, sig)
	}
	return nil
}

func mismatch(t string, v any) error {
	return fmt.Errorf("%w: %T does not fit signature %q", ErrBadMessage, v, t)
}

func (e *encoder) value(t string, v any) error {
	switch t[0] {
	case 'y':
		b, ok := v.(byte)
		if !ok {
			return mismatch(t, v)
		}
		e.buf = append(e.buf, b)
	case 'b':
		b, ok := v.(bool)
		if !ok {
			return mismatch(t, v)
		}
		if b {
			e.u32(1)
		} else {
			e.u32(0)
		}
	case 'i':
		n, ok := v.(int32)
		if !ok {
			return mismatch(t, v)
		}
		e.u32(uint32(n))
	case 'u':
		n, ok := v.(uint32)
		if !ok {
			return mismatch(t, v)
		}
		e.u32(n)
	case 's':
		s, ok := v.(string)
		if !ok {
			return mismatch(t, v)
		}
		e.str(s)
	case 'o':
		s, ok := v.(ObjectPath)
		if !ok {
			return mismatch(t, v)
		}
		e.str(string(s))
	case 'g':
		s, ok := v.(string)
		if !ok {
			return mismatch(t, v)
		}
		e.sig(s)
	case 'v':
		vr, ok := v.(Variant)
		if !ok {
			return mismatch(t, v)
		}
		e.sig(vr.Sig)
		return e.value(vr.Sig, vr.Value)
	case 'a':
		return e.array(t, v)
	case '(':
		fields, ok := v.([]any)
		if !ok {
			return mismatch(t, v)
		}
		e.align(8)
		return e.values(t[1:len(t)-1], fields)
	default:
		return fmt.Errorf("%w: unsupported type %q", ErrBadMessage, t)
	}
	return nil
}

func (e *encoder) array(t string, v any) error {
	elem := t[1:]
	e.u32(0)
	lenAt := len(e.buf) - 4
	e.align(alignment(elem[0]))
	start := len(e.buf)
	switch items := v.(type) {
	case []string:
		for _, s := range items {
			if err := e.value(elem, s); err != nil {
				return err
			}
		}
	case []any:
		for _, it := range items {
			if err := e.value(elem, it); err != nil {
				return err
			}
		}
	case map[string]Variant:
		if elem != "{sv}" {
			return mismatch(t, v)
		}
		for _, k := range sortedKeys(items) {
			e.align(8)
			e.str(k)
			if err := e.value("v", items[k]); err != nil {
				return err
			}
		}
	default:
		return mismatch(t, v)
	}
	binary.LittleEndian.PutUint32(e.buf[lenAt:], uint32(len(e.buf)-start))
	return nil
}

type decoder struct {
	buf   []byte
	pos   int
	order binary.ByteOrder
}

func (d *decoder) align(n int) {
	for d.pos%n != 0 {
		d.pos++
	}
}

func (d *decoder) take(n int) ([]byte, error) {
	if n < 0 || d.pos+n > len(d.buf) {
		return nil, fmt.Errorf("%w: truncated", ErrBadMessage)
	}
	b := d.buf[d.pos : d.pos+n]
	d.pos += n
	return b, nil
}

func (d *decoder) u32() (uint32, error) {
	d.align(4)
	b, err := d.take(4)
	if err != nil {
		return 0, err
	}
	return d.order.Uint32(b), nil
}

func (d *decoder) str() (string, error) {
	n, err := d.u32()
	if err != nil {
		return "", err
	}
	b, err := d.take(int(n) + 1)
	if err != nil {
		return "", err
	}
	return string(b[:n]), nil
}

func (d *decoder) sig() (string, error) {
	n, err := d.take(1)
	if err != nil {
		return "", err
	}
	b, err := d.take(int(n[0]) + 1)
	if err != nil {
		return "", err
	}
	return string(b[:n[0]]), nil
}

// values decodes every complete type of sig.
func (d *decoder) values(sig string) ([]any, error) {
	var out []any
	for sig != "" {
		t, rest, err := next(sig)
		if err != nil {
			return nil, err
		}
		v, err := d.value(t)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
		sig = rest
	}
	return out, nil
}

func (d *decoder) value(t string) (any, error) {
	switch t[0] {
	case 'y':
		b, err := d.take(1)
		if err != nil {
			return nil, err
		}
		return b[0], nil
	case 'b':
		n, err := d.u32()
		return n != 0, err
	case 'n', 'q':
		d.align(2)
		b, err := d.take(2)
		if err != nil {
			return nil, err
		}
		if t[0] == 'n' {
			return int16(d.order.Uint16(b)), nil
		}
		return d.order.Uint16(b), nil
	case 'i':
		n, err := d.u32()
		return int32(n), err
	case 'u':
		return d.u32()
	case 'x', 't', 'd':
		d.align(8)
		b, err := d.take(8)
		if err != nil {
			return nil, err
		}
		n := d.order.Uint64(b)
		switch t[0] {
		case 'x':
			return int64(n), nil
		case 'd':
			return math.Float64frombits(n), nil
		}
		return n, nil
	case 's':
		return d.str()
	case 'o':
		s, err := d.str()
		return ObjectPath(s), err
	case 'g':
		return d.sig()
	case 'v':
		s, err := d.sig()
		if err != nil {
			return nil, err
		}
		if _, rest, err := next(s); err != nil || rest != "" {
			return nil, fmt.Errorf("%w: variant signature %q", ErrBadMessage, s)
		}
		v, err := d.value(s)
		return Variant{Sig: s, Value: v}, err
	case 'a':
		n, err := d.u32()
		if err != nil {
			return nil, err
		}
		elem := t[1:]
		d.align(alignment(elem[0]))
		end := d.pos + int(n)
		if end > len(d.buf) {
			return nil, fmt.Errorf("%w: array overruns the message", ErrBadMessage)
		}
		items := []any{}
		for d.pos < end {
			before := d.pos
			v, err := d.value(elem)
			if err != nil {
				return nil, err
			}
			if d.pos == before {
				return nil, fmt.Errorf("%w: empty array element %q", ErrBadMessage, elem)
			}
			items = append(items, v)
		}
		return items, nil
	case '(', '{':
		d.align(8)
		return d.values(t[1 : len(t)-1])
	}
	return nil, fmt.Errorf("%w: unsupported type %q", ErrBadMessage, t)
}

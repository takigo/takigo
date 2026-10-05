package canvas

import (
	"github.com/takigo/takigo/color"
)

// ActiveFill sets -activefill.
func ActiveFill[C color.Spec](name C) ItemOption {
	return func(c *Canvas, item Item) error {
		col, err := c.ColorCache().Resolve(name)
		if err != nil {
			return err
		}
		if b := itemBase(item); b != nil {
			b.activeFill = col.Ref()
		}
		return nil
	}
}

// DisabledFill sets -disabledfill.
func DisabledFill[C color.Spec](name C) ItemOption {
	return func(c *Canvas, item Item) error {
		col, err := c.ColorCache().Resolve(name)
		if err != nil {
			return err
		}
		if b := itemBase(item); b != nil {
			b.disabledFill = col.Ref()
		}
		return nil
	}
}

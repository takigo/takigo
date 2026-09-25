package canvas

// ActiveFill sets -activefill.
func ActiveFill(name string) ItemOption {
	return func(c *Canvas, item Item) error {
		col, err := c.ColorCache().Get(name)
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
func DisabledFill(name string) ItemOption {
	return func(c *Canvas, item Item) error {
		col, err := c.ColorCache().Get(name)
		if err != nil {
			return err
		}
		if b := itemBase(item); b != nil {
			b.disabledFill = col.Ref()
		}
		return nil
	}
}

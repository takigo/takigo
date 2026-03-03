package platform

// DrawableFont extends the font.Font interface with platform-specific
// string drawing. This eliminates type assertions to *font.XftFont
// found across 20+ widget files.
type DrawableFont interface {
	// DrawString draws a string on a drawable at the given baseline position.
	// pixel is the color pixel value; r, g, b are 16-bit RGB components.
	DrawString(drawable DrawableID, x, y int, s string, pixel uint64, r, g, b uint16)
}

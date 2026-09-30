package window

import "testing"

func TestSetBackgroundPixelNoWindow(t *testing.T) {
	w := &Window{}
	w.SetBackgroundPixel(0xabcdef)
	if w.BackgroundPixel != 0xabcdef || w.serverBackground != 0 {
		t.Fatalf("BackgroundPixel = %x, serverBackground = %x", w.BackgroundPixel, w.serverBackground)
	}
}

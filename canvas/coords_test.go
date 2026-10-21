package canvas

import "testing"

// TestItemCoordsCopy checks that the coordinates an item hands out are a
// copy: a caller that writes into them must not change the item.
func TestItemCoordsCopy(t *testing.T) {
	c := newBenchCanvas()
	items := []Item{
		newRectOvalItem("rectangle", 1, 2, 3, 4, c),
		newRectOvalItem("oval", 1, 2, 3, 4, c),
		newArcItem(1, 2, 3, 4, c),
		newLineItem([]float64{1, 2, 3, 4}, c),
	}
	for _, item := range items {
		coords := item.Coords()
		if len(coords) == 0 {
			t.Fatalf("%s: no coordinates", item.Type())
		}
		want := coords[0]
		coords[0] = 99
		if got := item.Coords()[0]; got != want {
			t.Errorf("%s: Coords() handed out its own array: got %v, want %v", item.Type(), got, want)
		}
	}
}

// TestCanvasItemCoordsCopy checks the same through the canvas, which
// passes the item's coordinates out unchanged.
func TestCanvasItemCoordsCopy(t *testing.T) {
	c := newBenchCanvas()
	id := c.createItem(newRectOvalItem("rectangle", 1, 2, 3, 4, c), nil)
	coords := c.ItemCoords(id)
	if len(coords) != 4 {
		t.Fatalf("ItemCoords returned %v", coords)
	}
	coords[0] = 99
	if got := c.ItemCoords(id)[0]; got != 1 {
		t.Errorf("ItemCoords handed out the item's array: got %v, want 1", got)
	}
}

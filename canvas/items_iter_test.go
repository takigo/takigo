package canvas

import (
	"slices"
	"testing"
)

func TestItemsIterator(t *testing.T) {
	c := newBenchCanvas()
	a := c.CreateRectangle(0, 0, 10, 10, Tags("box"))
	b := c.CreateOval(0, 0, 10, 10)
	d := c.CreateRectangle(20, 20, 30, 30, Tags("box"))

	if got := slices.Collect(c.Items("box")); !slices.Equal(got, []ItemID{a, d}) {
		t.Errorf(`Items("box") = %v, want %v`, got, []ItemID{a, d})
	}
	if got := slices.Collect(c.Items(b)); !slices.Equal(got, []ItemID{b}) {
		t.Errorf("Items(id) = %v, want [%d]", got, b)
	}
	// Deleting while iterating is safe: the iteration is over a snapshot.
	n := 0
	for id := range c.Items("all") {
		c.Delete(id)
		n++
	}
	if n != 3 || len(c.FindWithTag("all")) != 0 {
		t.Errorf("visited %d items, %d left; want 3 and 0", n, len(c.FindWithTag("all")))
	}
}

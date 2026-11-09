package widget_test

import (
	"fmt"

	"github.com/takigo/takigo/widget"
)

// A Variable notifies its listeners when Set changes the value, as a Tcl
// variable trace does; widgets linked to it follow it the same way.
func ExampleVariable() {
	name := widget.NewVariable("")
	stop := name.OnChange(func(old, cur string) { fmt.Printf("%q -> %q\n", old, cur) })
	name.Set("Ada")
	name.Set("Ada") // unchanged: no notification
	name.Set("Grace")
	stop()
	name.Set("Linus")
	fmt.Println(name.Get())
	// Output:
	// "" -> "Ada"
	// "Ada" -> "Grace"
	// Linus
}

package option_test

import (
	"fmt"

	"github.com/takigo/takigo/option"
)

// The option values print as their Tk names.
func ExampleParseRelief() {
	for _, s := range []string{"raised", "sunken", "groove"} {
		fmt.Println(option.ParseRelief(s))
	}
	fmt.Println(option.AnchorNW, option.JustifyCenter, option.Vertical)
	fmt.Println(option.StickNSEW, option.StickN|option.StickW)
	// Output:
	// raised
	// sunken
	// groove
	// nw center vertical
	// nesw nw
}

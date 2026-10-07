package screenunit_test

import (
	"fmt"

	"github.com/takigo/takigo/screenunit"
)

// Distances are typed: points, millimetres, centimetres and inches convert
// to pixels at the screen's resolution, which the example pins to 96 dpi.
func ExampleParse() {
	screenunit.SetScreenDPI(1920, 508, 96)
	for _, s := range []string{"72p", "25.4m", "1c", "0.5i", "12", "huge"} {
		d, err := screenunit.Parse(s)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%s = %d px\n", s, d.Pixels())
	}
	fmt.Println(screenunit.Pt(3), screenunit.In(1).Pixels())
	// Output:
	// 72p = 96 px
	// 25.4m = 96 px
	// 1c = 38 px
	// 0.5i = 48 px
	// 12 = 12 px
	// screenunit: bad distance: "huge"
	// 3p 96
}

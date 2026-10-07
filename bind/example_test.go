package bind_test

import (
	"fmt"

	"github.com/takigo/takigo/bind"
)

// Parse reads Tk's event sequence syntax; a Sequence prints back in its
// canonical form.
func ExampleParse() {
	for _, s := range []string{"<Control-Shift-Key-s>", "<Double-Button-1>", "<<Copy>>", "a"} {
		seq, err := bind.Parse(s)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(seq)
	}
	// Output:
	// <Control-Shift-Key-s>
	// <Double-Button-1>
	// <<Copy>>
	// <Key-a>
}

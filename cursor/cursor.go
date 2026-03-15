// Package cursor defines platform-neutral cursor shapes.
// Each platform backend maps these to native cursor identifiers.
package cursor

// Shape represents a platform-neutral cursor shape.
type Shape uint

// Standard cursor shapes.
const (
	Arrow          Shape = iota // standard left pointer
	Crosshair                  // crosshair
	Fleur                      // move cursor (four-directional)
	Hand1                      // hand
	Hand2                      // pointing hand (for links)
	LeftPtr                    // default arrow (alias for Arrow on most platforms)
	Plus                       // plus sign
	QuestionArrow              // help cursor
	SBHDoubleArrow             // horizontal resize
	SBVDoubleArrow             // vertical resize
	SizingAngle                // corner resize
	TopLeftArrow               // top-left arrow
	Watch                      // busy/wait cursor
	XTerm                      // text insertion I-beam
	BottomRightCorner          // bottom-right corner resize
)

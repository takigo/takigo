// Package cursor defines standard cursor shapes.
// Shape values correspond to X11 cursor font indices (X11/cursorfont.h).
package cursor

// Standard cursor shapes.
const (
	Arrow        uint = 2   // XC_arrow — standard left pointer
	Crosshair    uint = 34  // XC_crosshair
	Fleur        uint = 52  // XC_fleur — move cursor
	Hand1        uint = 58  // XC_hand1
	Hand2        uint = 60  // XC_hand2 — pointing hand (for links)
	LeftPtr      uint = 68  // XC_left_ptr — default arrow
	Plus         uint = 90  // XC_plus
	QuestionArrow uint = 92 // XC_question_arrow — help cursor
	SBHDoubleArrow uint = 108 // XC_sb_h_double_arrow — horizontal resize
	SBVDoubleArrow uint = 116 // XC_sb_v_double_arrow — vertical resize
	SizingAngle  uint = 120 // XC_sizing — corner resize
	TopLeftArrow uint = 132 // XC_top_left_arrow
	Watch        uint = 150 // XC_watch — busy
	XTerm        uint = 152 // XC_xterm — text insertion
)

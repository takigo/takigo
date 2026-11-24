package dialog

import (
	"github.com/takigo/takigo/widget"

	tkimage "github.com/takigo/takigo/image"
)

// msgIcon returns the 32px message box icon for t: takigo's own drawings,
// or in a Classic App the ::tk::icons images of tk/library/icons.tcl that
// msgbox.tcl shows.
func msgIcon(app widget.AppContext, t MessageType) *tkimage.Photo {
	set := msgIcons
	if widget.Classic(app) {
		set = tkMsgIcons
	}
	src, ok := set[t]
	if !ok {
		src = set[MsgInfo]
	}
	return svgIcon(src, "")
}

const svg32 = `<svg width="32" height="32" version="1.1" xmlns="http://www.w3.org/2000/svg">`

// badge is a round 32px badge shaded from top to bottom.
func badge(top, bottom string) string {
	return `<linearGradient id="bg" x1="0" y1="1" x2="0" y2="31" gradientUnits="userSpaceOnUse">` +
		`<stop offset="0" stop-color="` + top + `"/><stop offset="1" stop-color="` + bottom + `"/></linearGradient>` +
		`<circle cx="16" cy="16" r="15" fill="url(#bg)"/>`
}

var msgIcons = map[MessageType]string{
	MsgError: svg32 + badge("#f2625e", "#d0312d") +
		`<path d="m11 11 10 10m0-10-10 10" stroke="#ffffff" stroke-width="3" stroke-linecap="round"/></svg>`,
	MsgWarning: svg32 +
		`<linearGradient id="bg" x1="0" y1="3" x2="0" y2="29" gradientUnits="userSpaceOnUse">` +
		`<stop offset="0" stop-color="#ffcd4d"/><stop offset="1" stop-color="#f59a0b"/></linearGradient>` +
		`<path d="m16 3.5 13.5 24h-27z" fill="url(#bg)" stroke="#f5a30f" stroke-width="3" stroke-linejoin="round"/>` +
		`<path d="m16 11.5v8" stroke="#3b2a05" stroke-width="3" stroke-linecap="round"/>` +
		`<circle cx="16" cy="24" r="1.75" fill="#3b2a05"/></svg>`,
	MsgInfo: svg32 + badge("#4ea5ff", "#1d6fdc") +
		`<circle cx="16" cy="9.5" r="2.1" fill="#ffffff"/>` +
		`<path d="m16 14.5v9" stroke="#ffffff" stroke-width="3.5" stroke-linecap="round"/></svg>`,
	MsgQuestion: svg32 + badge("#9282f2", "#5a42d4") +
		`<path d="m12 12.75a4 4 0 1 1 6.2 3.33c-1.3 0.86-2.2 1.6-2.2 3.17v0.5" fill="none" stroke="#ffffff" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>` +
		`<circle cx="16" cy="24.25" r="2" fill="#ffffff"/></svg>`,
}

// tkMsgIcons are ::tk::icons::error, warning, information and question
// (tk/library/icons.tcl, from the Vimix icon theme, CC BY-SA 4.0).
var tkMsgIcons = map[MessageType]string{
	MsgError: svg32 + `<circle cx="16" cy="16" r="16" fill="#d32f2f"/>` +
		`<g transform="rotate(45,16,16)" fill="#fff"><rect x="6" y="14" width="20" height="4"/>` +
		`<rect x="14" y="6" width="4" height="20"/></g></svg>`,
	MsgWarning: svg32 + `<circle cx="16" cy="16" r="16" fill="#f67400"/>` +
		`<circle cx="16" cy="24" r="2" fill="#fff"/><path d="m14 20h4v-14h-4z" fill="#fff"/></svg>`,
	MsgInfo: svg32 + `<circle cx="16" cy="16" r="16" fill="#2091df"/>` +
		`<circle cx="16" cy="8" r="2" fill="#fff"/><path d="m14 12h4v14h-4z" fill="#fff"/></svg>`,
	MsgQuestion: svg32 + `<circle cx="16" cy="16" r="16" fill="#5c6bc0"/>` +
		`<path d="m17.6 27.2h-3.2v-3.2h3.2zm3.312-12.4-1.44 1.472c-1.152 1.168-1.872 2.128-1.872 4.528h-3.2v-0.8c0-1.76 0.72-3.36 1.872-4.528l1.984-2.016a3.128 3.128 0 0 0 0.944-2.256c0-1.76-1.44-3.2-3.2-3.2s-3.2 1.44-3.2 3.2h-3.2c0-3.536 2.864-6.4 6.4-6.4s6.4 2.864 6.4 6.4c0 1.408-0.576 2.688-1.488 3.6z" fill="#fff"/></svg>`,
}

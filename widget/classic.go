package widget

import "github.com/takigo/takigo/window"

var classicKey = new(window.ValueKey)

// SetClassic marks the App whose root window is root as Classic: its
// widgets draw and behave exactly as Tk does, without the departures made
// for a better result (anti-aliased canvas items). takigo.Classic and the
// TAKIGO_CLASSIC environment variable set it; the demo comparison against
// Tk runs with it.
func SetClassic(root *window.Window, on bool) {
	root.SetValue(classicKey, on)
}

// Classic reports whether app is a Classic App; see SetClassic.
func Classic(app AppContext) bool {
	if app == nil || app.Window() == nil {
		return false
	}
	on, _ := app.Window().Value(classicKey).(bool)
	return on
}

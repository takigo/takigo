// Package takigo, in tests/, holds the black-box tests of the App: behaviour that
// crosses packages (appearance and themes, the clipboard, modal and nested
// loops, several Apps at once, pixel read-back of what widgets paint) and
// that needs a display. The helpers in internal/testutil start a private
// Xvfb for them; see AGENTS.md, "Tests".
package takigo

package bind

// defaultVirtualEvents defines the standard virtual event mappings.
// Each virtual event name maps to one or more physical event patterns.
var defaultVirtualEvents = map[string][]string{
	"Copy":      {"<Control-c>"},
	"Cut":       {"<Control-x>"},
	"Paste":     {"<Control-v>"},
	"SelectAll": {"<Control-a>"},
	"Undo":      {"<Control-z>"},
	"Redo":      {"<Control-y>", "<Control-Shift-z>"},
}

// installDefaultVirtualEvents registers the default virtual events on an engine.
func installDefaultVirtualEvents(e *Engine) {
	for name, patterns := range defaultVirtualEvents {
		for _, pat := range patterns {
			seq, err := Parse(pat)
			if err != nil {
				continue
			}
			e.virtualEvents[name] = append(e.virtualEvents[name], seq)
		}
	}
}

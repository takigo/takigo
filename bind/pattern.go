package bind

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
)

// Modifier represents modifier key flags in a binding pattern.
type Modifier uint32

const (
	ModShift Modifier = 1 << iota
	ModControl
	ModAlt
	ModMeta
	ModDouble
	ModTriple
)

// Pattern describes a single event pattern to match.
type Pattern struct {
	EventType event.Type
	Modifiers Modifier
	KeySym    platform.KeySym // 0 = any key
	Button    uint            // 0 = any button
	Virtual   string          // non-empty for virtual events like <<Copy>>
}

// Sequence is one or more patterns forming a complete binding specification.
type Sequence struct {
	Patterns []Pattern
}

// String returns the original pattern string representation.
func (s Sequence) String() string {
	var parts []string
	for _, p := range s.Patterns {
		parts = append(parts, patternString(p))
	}
	return strings.Join(parts, "")
}

func patternString(p Pattern) string {
	if p.Virtual != "" {
		return "<<" + p.Virtual + ">>"
	}
	var b strings.Builder
	b.WriteByte('<')
	if p.Modifiers&ModDouble != 0 {
		b.WriteString("Double-")
	}
	if p.Modifiers&ModTriple != 0 {
		b.WriteString("Triple-")
	}
	if p.Modifiers&ModControl != 0 {
		b.WriteString("Control-")
	}
	if p.Modifiers&ModShift != 0 {
		b.WriteString("Shift-")
	}
	if p.Modifiers&ModAlt != 0 {
		b.WriteString("Alt-")
	}
	if p.Modifiers&ModMeta != 0 {
		b.WriteString("Meta-")
	}
	b.WriteString(eventTypeName(p.EventType))
	if p.Button != 0 {
		b.WriteByte('-')
		b.WriteString(strconv.Itoa(int(p.Button)))
	}
	if p.KeySym != 0 {
		b.WriteByte('-')
		b.WriteString(keySymName(p.KeySym))
	}
	b.WriteByte('>')
	return b.String()
}

func eventTypeName(t event.Type) string {
	switch t {
	case event.KeyPressType:
		return "Key"
	case event.KeyReleaseType:
		return "KeyRelease"
	case event.ButtonPressType:
		return "Button"
	case event.ButtonReleaseType:
		return "ButtonRelease"
	case event.MotionType:
		return "Motion"
	case event.EnterType:
		return "Enter"
	case event.LeaveType:
		return "Leave"
	case event.FocusInType:
		return "FocusIn"
	case event.FocusOutType:
		return "FocusOut"
	case event.ExposeType:
		return "Expose"
	case event.ConfigureType:
		return "Configure"
	case event.DestroyType:
		return "Destroy"
	case event.MapType:
		return "Map"
	case event.UnmapType:
		return "Unmap"
	default:
		return "Unknown"
	}
}

// Parse parses a Tk-style event pattern string into a Sequence.
// Supported formats:
//   - <Button-1>, <ButtonRelease-1>
//   - <Key-a>, <Key-Return>, <KeyRelease-Escape>
//   - <Control-a>, <Control-Shift-x>
//   - <Double-Button-1>, <Triple-Button-1>
//   - <Motion>, <Enter>, <Leave>, <FocusIn>, <FocusOut>
//   - <Configure>, <Expose>, <Destroy>, <Map>, <Unmap>
//   - <<VirtualName>> (virtual events)
//   - Single character shorthand: "a" = <Key-a>
func Parse(pattern string) (Sequence, error) {
	if len(pattern) == 0 {
		return Sequence{}, fmt.Errorf("bind: empty pattern")
	}

	// Single character shorthand.
	if len(pattern) == 1 {
		ch := pattern[0]
		ks := lookupKeySym(pattern)
		if ks == 0 {
			return Sequence{}, fmt.Errorf("bind: unknown key %q", ch)
		}
		return Sequence{Patterns: []Pattern{{
			EventType: event.KeyPressType,
			KeySym:    ks,
		}}}, nil
	}

	var patterns []Pattern
	i := 0
	for i < len(pattern) {
		if pattern[i] != '<' {
			return Sequence{}, fmt.Errorf("bind: expected '<' at position %d in %q", i, pattern)
		}

		// Find matching '>'.
		end := strings.IndexByte(pattern[i:], '>')
		if end < 0 {
			return Sequence{}, fmt.Errorf("bind: unmatched '<' in %q", pattern)
		}
		end += i

		inner := pattern[i+1 : end]

		// Virtual event: <<Name>>
		if strings.HasPrefix(inner, "<") && end+1 < len(pattern) && pattern[end+1] == '>' {
			vname := inner[1:]
			if vname == "" {
				return Sequence{}, fmt.Errorf("bind: empty virtual event name")
			}
			patterns = append(patterns, Pattern{Virtual: vname})
			i = end + 2
			continue
		}

		p, err := parseInner(inner)
		if err != nil {
			return Sequence{}, fmt.Errorf("bind: %w in %q", err, pattern)
		}
		patterns = append(patterns, p)
		i = end + 1
	}

	if len(patterns) == 0 {
		return Sequence{}, fmt.Errorf("bind: no patterns in %q", pattern)
	}

	return Sequence{Patterns: patterns}, nil
}

// MustParse is like Parse but panics on error.
func MustParse(pattern string) Sequence {
	s, err := Parse(pattern)
	if err != nil {
		panic(err)
	}
	return s
}

// parseInner parses the content between < and >.
func parseInner(s string) (Pattern, error) {
	parts := strings.Split(s, "-")
	var p Pattern
	var mods Modifier

	// Process each part. The last part(s) determine the event type and detail.
	// Modifiers come first, then event type, then optional detail.
	for len(parts) > 0 {
		part := parts[0]
		lower := toLower(part)

		// Try as modifier.
		if mod, ok := parseModifier(lower); ok {
			mods |= mod
			parts = parts[1:]
			continue
		}

		// Try as event type.
		if et, ok := parseEventType(lower); ok {
			p.EventType = et
			parts = parts[1:]

			// Remaining part is detail (button number or key name).
			if len(parts) > 0 {
				detail := strings.Join(parts, "-")
				if err := setDetail(&p, detail); err != nil {
					return Pattern{}, err
				}
				parts = nil
			}
			break
		}

		// If we reach here, treat everything remaining as a key name.
		keyName := strings.Join(parts, "-")
		p.EventType = event.KeyPressType
		ks := lookupKeySym(keyName)
		if ks == 0 {
			return Pattern{}, fmt.Errorf("unknown key or event type %q", keyName)
		}
		p.KeySym = ks
		parts = nil
	}

	if p.EventType == 0 && p.Virtual == "" {
		return Pattern{}, fmt.Errorf("no event type specified")
	}

	p.Modifiers = mods
	return p, nil
}

// parseModifier checks if a token is a modifier keyword.
func parseModifier(lower string) (Modifier, bool) {
	switch lower {
	case "shift":
		return ModShift, true
	case "control", "ctrl":
		return ModControl, true
	case "alt", "mod1":
		return ModAlt, true
	case "meta", "mod4", "super":
		return ModMeta, true
	case "double":
		return ModDouble, true
	case "triple":
		return ModTriple, true
	}
	return 0, false
}

// parseEventType checks if a token is an event type keyword.
func parseEventType(lower string) (event.Type, bool) {
	switch lower {
	case "key", "keypress":
		return event.KeyPressType, true
	case "keyrelease":
		return event.KeyReleaseType, true
	case "button", "buttonpress":
		return event.ButtonPressType, true
	case "buttonrelease":
		return event.ButtonReleaseType, true
	case "motion":
		return event.MotionType, true
	case "enter":
		return event.EnterType, true
	case "leave":
		return event.LeaveType, true
	case "focusin":
		return event.FocusInType, true
	case "focusout":
		return event.FocusOutType, true
	case "expose":
		return event.ExposeType, true
	case "configure":
		return event.ConfigureType, true
	case "destroy":
		return event.DestroyType, true
	case "map":
		return event.MapType, true
	case "unmap":
		return event.UnmapType, true
	}
	return 0, false
}

// setDetail sets the detail (button number or keysym) on a pattern.
func setDetail(p *Pattern, detail string) error {
	switch p.EventType {
	case event.ButtonPressType, event.ButtonReleaseType:
		n, err := strconv.Atoi(detail)
		if err != nil || n < 1 || n > 5 {
			return fmt.Errorf("invalid button number %q", detail)
		}
		p.Button = uint(n)
	case event.KeyPressType, event.KeyReleaseType:
		ks := lookupKeySym(detail)
		if ks == 0 {
			return fmt.Errorf("unknown key %q", detail)
		}
		p.KeySym = ks
	default:
		return fmt.Errorf("detail %q not valid for event type", detail)
	}
	return nil
}

// matches checks if a pattern matches an event.
func (p *Pattern) matches(ev *event.Event, clickMods Modifier) bool {
	if p.Virtual != "" {
		return false // virtual events matched separately
	}

	// Event type must match.
	if p.EventType != ev.Type {
		return false
	}

	// Check modifiers against event state.
	evMods := eventModifiers(ev)
	evMods |= clickMods // add double/triple from click tracking

	reqMods := p.Modifiers &^ (ModDouble | ModTriple) // keyboard/pointer modifiers
	if reqMods&ModControl != 0 && evMods&ModControl == 0 {
		return false
	}
	if reqMods&ModShift != 0 && evMods&ModShift == 0 {
		return false
	}
	if reqMods&ModAlt != 0 && evMods&ModAlt == 0 {
		return false
	}
	if reqMods&ModMeta != 0 && evMods&ModMeta == 0 {
		return false
	}

	// Check double/triple.
	if p.Modifiers&ModDouble != 0 && clickMods&ModDouble == 0 {
		return false
	}
	if p.Modifiers&ModTriple != 0 && clickMods&ModTriple == 0 {
		return false
	}

	// Check detail.
	if p.Button != 0 && p.Button != ev.Button {
		return false
	}
	if p.KeySym != 0 && p.KeySym != ev.KeySym {
		return false
	}

	return true
}

// specificity of a sequence is the sum over its patterns, so a completed
// multi-event sequence outranks its single-event suffix.
func (s Sequence) specificity() int {
	n := 0
	for i := range s.Patterns {
		n += s.Patterns[i].specificity()
	}
	return n
}

// specificity returns a score for pattern specificity (higher = more specific).
// More specific patterns take priority when multiple match.
func (p *Pattern) specificity() int {
	score := 0
	if p.KeySym != 0 || p.Button != 0 {
		score += 4
	}
	mods := p.Modifiers &^ (ModDouble | ModTriple)
	for mods != 0 {
		score += 2
		mods &= mods - 1 // clear lowest set bit
	}
	if p.Modifiers&ModDouble != 0 {
		score++
	}
	if p.Modifiers&ModTriple != 0 {
		score++
	}
	return score
}

// eventModifiers extracts our Modifier flags from platform state.
func eventModifiers(ev *event.Event) Modifier {
	var m Modifier
	if ev.State&platform.ControlMask != 0 {
		m |= ModControl
	}
	if ev.State&platform.ShiftMask != 0 {
		m |= ModShift
	}
	if ev.State&platform.Mod1Mask != 0 {
		m |= ModAlt
	}
	if ev.State&platform.Mod4Mask != 0 {
		m |= ModMeta
	}
	return m
}

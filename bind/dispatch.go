package bind

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

const (
	// doubleClickMs is the maximum time between clicks for double-click.
	doubleClickMs = 500
)

// tagInfo holds the binding tag chain for a registered window.
type tagInfo struct {
	win       *window.Window
	className string
	tags      []string // binding tag chain: [pathName, className, toplevelPath, "all"]
}

// Engine is the binding engine that sits above the event dispatcher.
// It provides Tk-style tag-based event binding with pattern matching,
// virtual events, and double-click detection.
type Engine struct {
	table      *BindingTable
	display    *window.Display
	dispatcher *event.Dispatcher

	// Per-window tag info.
	tags map[platform.WindowID]*tagInfo

	// Virtual event definitions: name → physical sequences, plus the
	// names in definition order so matching is deterministic.
	virtualEvents map[string][]Sequence
	virtualOrder  []string

	// prom holds multi-event bindings whose leading patterns have matched
	// and that wait for their next pattern (tkBind.c's promotion lists).
	prom []promEntry

	// Double-click state tracking.
	lastClickTime platform.Timestamp
	lastClickWin  platform.WindowID
	lastClickBtn  uint
	clickCount    int
}

// tagInfoFor returns w's tag chain, creating the default one (Tk's
// bindtags: path, class, toplevel, "all") on first use, when the widget
// constructor has set w.Class.
func (e *Engine) tagInfoFor(w *window.Window) *tagInfo {
	if w == nil || w.PlatformID == 0 {
		return nil
	}
	info := e.tags[w.PlatformID]
	if info == nil {
		info = &tagInfo{win: w, className: w.Class, tags: buildTagChain(w, w.Class)}
		e.tags[w.PlatformID] = info
	}
	return info
}

// dispatch is the binding-tag chain (event.ChainFunc) run for every event.
// It resolves the window, computes click modifiers, walks the tag chain,
// matches patterns, and calls handlers; with runClass it runs the window's
// own handlers when it reaches the class tag. It reports false for windows
// it does not know.
func (e *Engine) dispatch(ev *event.Event, runClass bool) bool {
	w := e.display.LookupWindow(ev.Window)
	info := e.tagInfoFor(w)
	if info == nil {
		return false
	}

	// Compute double/triple click modifiers for button press events.
	clickMods := e.updateClickState(ev)

	completed := e.advancePromoted(ev, clickMods)

	// Walk the tag chain and dispatch. Like Tk_BindEvent, stop once a
	// binding has destroyed the window.
	for _, tag := range info.tags {
		if runClass && tag == info.className {
			runClass = false
			e.dispatcher.DispatchWindow(ev)
			if w.IsDestroyed() {
				return true
			}
		}
		bindings := e.table.Lookup(tag)

		// Find best matching binding (most specific pattern wins); a
		// completed multi-event sequence competes on the sum of its
		// patterns' specificity, so the longer match wins (IsBetterMatch).
		var bestBinding *binding
		bestScore := -1
		for i := range bindings {
			b := &bindings[i]
			n := len(b.seq.Patterns)
			if n == 0 {
				continue
			}
			pat := &b.seq.Patterns[0]
			if !pat.matches(ev, clickMods) {
				continue
			}
			if n > 1 {
				e.promote(tag, b.seq, 1, ev.Window)
				continue
			}
			if score := pat.specificity(); score > bestScore {
				bestScore = score
				bestBinding = b
			}
		}
		for _, c := range completed {
			if c.tag != tag {
				continue
			}
			if b := findBinding(bindings, c.key); b != nil && c.score > bestScore {
				bestScore = c.score
				bestBinding = b
			}
		}

		// Also check virtual events: if the physical event matches a virtual
		// definition, try to dispatch bindings tagged with that virtual name.
		if bestBinding == nil {
			bestBinding = e.matchVirtual(ev, clickMods, bindings)
		}

		if bestBinding != nil {
			ed := &EventData{RawEvent: ev}
			if bestBinding.handler(ed) || w.IsDestroyed() {
				return true // break chain
			}
		}
	}
	return true
}

// promEntry is a multi-event binding of tag waiting, on window, for its
// pattern number next.
type promEntry struct {
	tag    string
	key    string // Sequence.String(), to find the binding again
	seq    Sequence
	next   int
	window platform.WindowID
}

// completedSeq is a multi-event binding whose last pattern just matched.
type completedSeq struct {
	tag   string
	key   string
	score int
}

// promote records that seq (bound to tag) matched up to pattern next-1 on
// window, unless it is already waiting there.
func (e *Engine) promote(tag string, seq Sequence, next int, window platform.WindowID) {
	key := seq.String()
	for _, p := range e.prom {
		if p.tag == tag && p.key == key && p.next == next && p.window == window {
			return
		}
	}
	e.prom = append(e.prom, promEntry{tag: tag, key: key, seq: seq, next: next, window: window})
}

// advancePromoted matches ev against every waiting sequence, as
// Tk_BindEvent does with its promotion lists: a matching entry completes
// or moves to its next pattern; a non-matching one stays unless the event
// rules it out.
func (e *Engine) advancePromoted(ev *event.Event, clickMods Modifier) []completedSeq {
	if len(e.prom) == 0 {
		return nil
	}
	var completed []completedSeq
	kept := e.prom[:0:0]
	for _, p := range e.prom {
		if p.window != ev.Window {
			continue
		}
		pat := &p.seq.Patterns[p.next]
		if pat.matches(ev, clickMods) {
			if p.next == len(p.seq.Patterns)-1 {
				completed = append(completed, completedSeq{tag: p.tag, key: p.key, score: p.seq.specificity()})
			} else {
				p.next++
				kept = append(kept, p)
			}
			continue
		}
		if !promotionSurvives(pat, ev) {
			continue
		}
		kept = append(kept, p)
	}
	e.prom = kept
	return completed
}

// promotionSurvives reports whether a sequence waiting for pat outlives
// ev, which did not match it. Following Tk_BindEvent's expiry rules, it is
// dropped by an event of pat's type with a different detail (another key
// or button) and by a switch between key and button events; modifier key
// presses and releases, and other event types, pass through.
func promotionSurvives(pat *Pattern, ev *event.Event) bool {
	isKey := ev.Type == event.KeyPressType || ev.Type == event.KeyReleaseType
	isButton := ev.Type == event.ButtonPressType || ev.Type == event.ButtonReleaseType
	if isKey && isModifierKeySym(ev.KeySym) {
		return true
	}
	if ev.Type == event.KeyReleaseType && pat.EventType != event.KeyReleaseType {
		return true
	}
	if pat.EventType == ev.Type {
		if pat.KeySym != 0 && isKey && ev.KeySym != pat.KeySym {
			return false
		}
		if pat.Button != 0 && isButton && ev.Button != pat.Button {
			return false
		}
	}
	patKey := pat.EventType == event.KeyPressType || pat.EventType == event.KeyReleaseType
	patButton := pat.EventType == event.ButtonPressType || pat.EventType == event.ButtonReleaseType
	return !(patKey && isButton) && !(patButton && isKey)
}

// isModifierKeySym reports whether ks is a modifier key (Shift, Control,
// Caps/Shift Lock, Meta, Alt, Super, Hyper, AltGr/Mode_switch).
func isModifierKeySym(ks platform.KeySym) bool {
	return (ks >= 0xffe1 && ks <= 0xffee) || ks == 0xff7e || ks == 0xfe03
}

// findBinding returns the binding in bindings whose sequence prints as key.
func findBinding(bindings []binding, key string) *binding {
	for i := range bindings {
		if bindings[i].seq.String() == key {
			return &bindings[i]
		}
	}
	return nil
}

// matchVirtual returns the binding for a virtual event whose physical
// definition matches ev. When several do, the most specific physical
// pattern wins and ties go to the earliest-defined virtual event, so the
// choice does not depend on map iteration order.
func (e *Engine) matchVirtual(ev *event.Event, clickMods Modifier, bindings []binding) *binding {
	var best *binding
	bestScore := -1
	for _, vname := range e.virtualOrder {
		score := -1
		for _, vs := range e.virtualEvents[vname] {
			if len(vs.Patterns) == 1 && vs.Patterns[0].matches(ev, clickMods) {
				score = max(score, vs.Patterns[0].specificity())
			}
		}
		if score <= bestScore {
			continue
		}
		for i := range bindings {
			b := &bindings[i]
			if len(b.seq.Patterns) == 1 && b.seq.Patterns[0].Virtual == vname {
				best, bestScore = b, score
				break
			}
		}
	}
	return best
}

// updateClickState tracks button press timing for double/triple click.
// Returns ModDouble or ModTriple modifier if applicable.
func (e *Engine) updateClickState(ev *event.Event) Modifier {
	if ev.Type != event.ButtonPressType {
		return 0
	}

	elapsed := uint64(ev.Time) - uint64(e.lastClickTime)
	if e.lastClickWin == ev.Window &&
		e.lastClickBtn == ev.Button &&
		elapsed < doubleClickMs {
		e.clickCount++
	} else {
		e.clickCount = 1
	}

	e.lastClickTime = ev.Time
	e.lastClickWin = ev.Window
	e.lastClickBtn = ev.Button

	switch e.clickCount {
	case 2:
		return ModDouble
	case 3:
		return ModTriple
	default:
		return 0
	}
}

// buildTagChain creates the default tag chain for a window:
// [pathName, className, toplevelPath, "all"]
func buildTagChain(w *window.Window, className string) []string {
	tags := []string{w.PathName, className}

	// Find the toplevel window.
	tlPath := findToplevelPath(w)
	if tlPath != w.PathName {
		tags = append(tags, tlPath)
	}

	tags = append(tags, "all")
	return tags
}

// findToplevelPath walks up the parent chain to find the toplevel window path.
func findToplevelPath(w *window.Window) string {
	for w != nil {
		if w.Flags&window.FlagTopLevel != 0 || w.Parent == nil {
			return w.PathName
		}
		w = w.Parent
	}
	return "."
}

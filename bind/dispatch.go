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
	table   *BindingTable
	display *window.Display

	// Per-window tag info.
	tags map[platform.WindowID]*tagInfo

	// Virtual event definitions: name → physical sequences, plus the
	// names in definition order so matching is deterministic.
	virtualEvents map[string][]Sequence
	virtualOrder  []string

	// Double-click state tracking.
	lastClickTime platform.Timestamp
	lastClickWin  platform.WindowID
	lastClickBtn  uint
	clickCount    int
}

// dispatch is the core event dispatch function called for every event.
// It resolves the window, computes click modifiers, walks the tag chain,
// matches patterns, and calls handlers.
func (e *Engine) dispatch(ev *event.Event) {
	w := e.display.LookupWindow(ev.Window)
	if w == nil {
		return
	}

	info := e.tags[ev.Window]
	if info == nil {
		return
	}

	// Compute double/triple click modifiers for button press events.
	clickMods := e.updateClickState(ev)

	// Walk the tag chain and dispatch.
	for _, tag := range info.tags {
		bindings := e.table.Lookup(tag)
		if len(bindings) == 0 {
			continue
		}

		// Find best matching binding (most specific pattern wins).
		var bestBinding *binding
		bestScore := -1

		for i := range bindings {
			b := &bindings[i]
			if len(b.seq.Patterns) != 1 {
				continue // only single-pattern sequences for now
			}
			pat := &b.seq.Patterns[0]

			if pat.matches(ev, clickMods) {
				score := pat.specificity()
				if score > bestScore {
					bestScore = score
					bestBinding = b
				}
			}
		}

		// Also check virtual events: if the physical event matches a virtual
		// definition, try to dispatch bindings tagged with that virtual name.
		if bestBinding == nil {
			bestBinding = e.matchVirtual(ev, clickMods, bindings)
		}

		if bestBinding != nil {
			ed := &EventData{RawEvent: ev}
			if bestBinding.handler(ed) {
				return // break chain
			}
		}
	}
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

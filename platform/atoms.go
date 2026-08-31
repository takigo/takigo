package platform

// Atoms holds the well-known atom IDs a backend resolves once during
// startup. Each platform backend populates an *Atoms value during
// NewDisplayServer and exposes it via DisplayServer.Atoms so callers
// can compare against X11-style constants (WM_NAME, PRIMARY, ...)
// without going through package-level mutable globals.
//
// Field names follow X11 conventions so call sites read like
// atoms.Primary, atoms.WMName, etc.
type Atoms struct {
	WMName        AtomID
	String        AtomID
	WMNormalHints AtomID
	Primary       AtomID
	Secondary     AtomID
	Atom          AtomID
	Cardinal      AtomID
	Window        AtomID
}

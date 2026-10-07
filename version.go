package takigo

import (
	"runtime/debug"
	"sync"
)

const modulePath = "github.com/takigo/takigo"

// Version reports the takigo module version the program was built with, as
// recorded by the go command: a release tag such as "v0.2.0", a
// pseudo-version for an untagged commit, or "(devel)" when the version is
// unknown (a replace directive pointing at a directory, go run on a file).
// The version is not compiled in, so it cannot disagree with the tags.
func Version() string { return version() }

var version = sync.OnceValue(func() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	return moduleVersion(bi)
})

func moduleVersion(bi *debug.BuildInfo) string {
	m := &bi.Main
	if m.Path != modulePath {
		m = nil
		for _, d := range bi.Deps {
			if d.Path == modulePath {
				m = d
				break
			}
		}
	}
	if m != nil && m.Replace != nil {
		m = m.Replace
	}
	if m == nil || m.Version == "" {
		return "(devel)"
	}
	return m.Version
}

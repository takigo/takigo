package takigo

import (
	"runtime/debug"
	"testing"
)

func TestModuleVersion(t *testing.T) {
	dep := func(path, ver string, repl *debug.Module) *debug.Module {
		return &debug.Module{Path: path, Version: ver, Replace: repl}
	}
	tests := []struct {
		name string
		bi   debug.BuildInfo
		want string
	}{
		{"main module", debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "v0.2.0"}}, "v0.2.0"},
		{"main pseudo-version", debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "v0.2.1-0.20261007120000-abcdef123456+dirty"}},
			"v0.2.1-0.20261007120000-abcdef123456+dirty"},
		{"main without version", debug.BuildInfo{Main: debug.Module{Path: modulePath}}, "(devel)"},
		{"dependency", debug.BuildInfo{
			Main: debug.Module{Path: "example.com/app", Version: "v1.0.0"},
			Deps: []*debug.Module{dep("example.com/other", "v9.9.9", nil), dep(modulePath, "v0.3.0", nil)},
		}, "v0.3.0"},
		{"replaced by a version", debug.BuildInfo{
			Main: debug.Module{Path: "example.com/app"},
			Deps: []*debug.Module{dep(modulePath, "v0.3.0", dep("example.com/fork", "v0.3.1", nil))},
		}, "v0.3.1"},
		{"replaced by a directory", debug.BuildInfo{
			Main: debug.Module{Path: "example.com/app"},
			Deps: []*debug.Module{dep(modulePath, "v0.3.0", dep("../takigo", "", nil))},
		}, "(devel)"},
		{"not linked", debug.BuildInfo{Main: debug.Module{Path: "example.com/app"}}, "(devel)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := moduleVersion(&tt.bi); got != tt.want {
				t.Errorf("moduleVersion = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	if Version() == "" {
		t.Error("Version() is empty")
	}
}

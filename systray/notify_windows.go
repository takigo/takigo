//go:build windows

package systray

func notify(string, string) error { return ErrUnsupported }

//go:build (linux || freebsd || openbsd || netbsd) && !cgo

package takigo

import (
	"errors"

	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
)

// platformInit without cgo: the X11 backend binds Xlib through cgo, so
// only the display-free packages work in such a build.
func platformInit(string) (platform.DisplayServer, platform.EventParser, font.FontOpener, error) {
	return nil, nil, nil, errors.New("takigo: the X11 backend needs cgo (CGO_ENABLED=1)")
}

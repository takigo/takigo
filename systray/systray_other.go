//go:build !(linux || freebsd || openbsd || netbsd)

package systray

import (
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// New reports ErrUnsupported: only the X11 tray protocol is implemented.
func New(app widget.AppContext, winDisplay *window.Display, opts ...TrayOption) (*TrayIcon, error) {
	return nil, ErrUnsupported
}

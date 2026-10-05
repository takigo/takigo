//go:build !(linux || freebsd || openbsd || netbsd)

package systray

import (
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// New reports ErrUnsupported: only the X11 tray protocol is implemented.
func New(app widget.AppContext, winDisplay *window.Display, opts ...TrayOption) (*TrayIcon, error) {
	return nil, ErrUnsupported
}

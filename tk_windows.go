//go:build windows

package takigo

import (
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/platform"
	winplatform "github.com/takigo/takigo/platform/windows"
)

// platformInit creates the platform-specific display server, event parser,
// and font opener. On Windows, this uses Win32/GDI APIs.
func platformInit(displayName string) (platform.DisplayServer, platform.EventParser, font.FontOpener, error) {
	server, fontOpener, err := winplatform.NewDisplayServer(displayName)
	if err != nil {
		return nil, nil, nil, err
	}

	parser := server.EventParser()
	return server, parser, fontOpener, nil
}

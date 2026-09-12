//go:build windows

package takigo

import (
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
	winplatform "github.com/msorc/takigo/platform/windows"
)

// platformInit creates the platform-specific display server, event parser,
// and font opener. On Windows, this uses Win32/GDI APIs.
func platformInit(displayName string) (platform.DisplayServer, platform.EventParser, font.FontOpener, error) {
	server, err := winplatform.NewDisplayServer(displayName)
	if err != nil {
		return nil, nil, nil, err
	}

	parser := server.EventParser()
	var fontOpener font.FontOpener
	if core, ok := server.(interface{ FontOpener(int) font.FontOpener }); ok {
		fontOpener = core.FontOpener(server.DefaultScreen())
	}

	return server, parser, fontOpener, nil
}

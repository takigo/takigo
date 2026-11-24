package systray

import (
	goimage "image"
	"testing"

	tkimage "github.com/takigo/takigo/image"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

type modifyServer struct {
	platform.DisplayServer
	icons []*platform.IconImage
}

func (s *modifyServer) ModifyTrayIcon(_ platform.WindowID, icon *platform.IconImage, _ string) {
	s.icons = append(s.icons, icon)
}

// TestConfigureSamePhoto checks that Configure with the photo the icon
// already shows sends its pixels again, so a changed photo shows, and that
// a Configure without TrayImage leaves the icon alone.
func TestConfigureSamePhoto(t *testing.T) {
	srv := &modifyServer{}
	photo := tkimage.NewPhoto("p", goimage.NewRGBA(goimage.Rect(0, 0, 4, 4)))
	tray := &TrayIcon{display: &window.Display{Server: srv}, win: 1, native: true, image: photo}

	tray.Configure(TrayImage(photo))
	tray.Configure(TrayTooltip("tip"))
	if len(srv.icons) != 2 || srv.icons[0] == nil || srv.icons[0].Width != 4 {
		t.Fatalf("Configure(TrayImage(same photo)) sent %v, want its pixels", srv.icons)
	}
	if srv.icons[1] != nil {
		t.Error("a tooltip change re-sent the image")
	}
}

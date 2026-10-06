package image

import (
	goimage "image"
	"testing"

	"github.com/takigo/takigo/platform"
)

// countingServer records pixmap traffic; other methods are unused.
type countingServer struct {
	platform.DisplayServer
	next    platform.PixmapID
	uploads int
	live    map[platform.PixmapID]bool
}

func (s *countingServer) CreatePixmap(platform.DrawableID, uint, uint, uint) platform.PixmapID {
	s.next++
	s.live[s.next] = true
	return s.next
}
func (s *countingServer) FreePixmap(id platform.PixmapID) { delete(s.live, id) }
func (s *countingServer) PutImageRGBA(platform.DrawableID, platform.GCID, int, []byte, int, int, int, int, int, int, int, int, int, uint64) {
	s.uploads++
}
func (s *countingServer) CopyArea(platform.DrawableID, platform.DrawableID, platform.GCID, int, int, uint, uint, int, int) {
}

func newTestPhoto(alpha uint8) *Photo {
	img := goimage.NewRGBA(goimage.Rect(0, 0, 4, 4))
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = alpha
	}
	return NewPhoto("p", img)
}

func TestPhotoPixmapCache(t *testing.T) {
	draw := func(p *Photo, s *countingServer, bg uint64) {
		p.Draw(s, 1, platform.GCID(0), 24, 0, 0, 4, 4, 0, 0, bg)
	}

	s := &countingServer{live: map[platform.PixmapID]bool{}}
	opaque := newTestPhoto(0xff)
	for _, bg := range []uint64{1, 2, 3, 1, 2, 3} {
		draw(opaque, s, bg)
	}
	if s.uploads != 1 {
		t.Errorf("opaque photo uploaded %d times across backgrounds, want 1", s.uploads)
	}

	s = &countingServer{live: map[platform.PixmapID]bool{}}
	clearPh := newTestPhoto(0x80)
	for range 10 {
		draw(clearPh, s, 1)
		draw(clearPh, s, 2)
	}
	if s.uploads != 2 {
		t.Errorf("translucent photo on two backgrounds uploaded %d times, want 2", s.uploads)
	}
	for bg := range uint64(maxPhotoPixmaps + 2) {
		draw(clearPh, s, 10+bg)
	}
	if len(s.live) != maxPhotoPixmaps {
		t.Errorf("%d pixmaps live, want at most %d", len(s.live), maxPhotoPixmaps)
	}
	clearPh.Invalidate()
	draw(clearPh, s, 1)
	clearPh.Destroy()
	if len(s.live) != 0 {
		t.Errorf("%d pixmaps leaked after Destroy", len(s.live))
	}
}

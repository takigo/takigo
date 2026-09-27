package busy

import (
	"testing"

	"github.com/msorc/takigo/platform"
)

func TestBusyWinNil(t *testing.T) {
	var b *BusyWin
	b.Release() // Should not panic
}

func TestBusyWinZeroOverlay(t *testing.T) {
	b := &BusyWin{overlay: platform.WindowID(0)}
	b.Release() // Should not panic
}

func TestBusyWinEmpty(t *testing.T) {
	b := &BusyWin{}
	b.Release() // Should not panic
}

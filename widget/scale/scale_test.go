package scale

import (
	"slices"
	"testing"

	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/window"
)

type fixedFont struct{}

func (fixedFont) Attrs() font.Attributes { return font.Attributes{} }
func (fixedFont) Metrics() font.Metrics {
	return font.Metrics{Ascent: 8, Descent: 2, MaxWidth: 6}
}
func (fixedFont) MeasureString(s string) int { return 6 * len(s) }
func (fixedFont) Close()                     {}

func TestTicks(t *testing.T) {
	tests := []struct {
		name                string
		from, to, res, tick float64
		want                []float64
	}{
		{"plain", 0, 4, 1, 1, []float64{0, 1, 2, 3, 4}},
		{"below half resolution", 0, 10, 1, 0.25, nil},
		{"rounds up to resolution", 0, 3, 1, 0.5, []float64{0, 1, 2, 3}},
		{"sign follows from-to", 4, 0, 1, 2, []float64{4, 2, 0}},
		{"negative interval", 0, 4, 1, -2, []float64{0, 2, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scale{From: tt.from, To: tt.to, Resolution: tt.res, TickInterval: tt.tick}
			s.Win = &window.Window{Width: 1000, Height: 1000}
			s.Font = fixedFont{}
			if got := s.ticks(); !slices.Equal(got, tt.want) {
				t.Errorf("ticks() = %v, want %v", got, tt.want)
			}
		})
	}
}

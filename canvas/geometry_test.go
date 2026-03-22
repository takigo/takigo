package canvas

import (
	"math"
	"testing"
)

// --- rectPointDistance ---

func TestRectPointDistanceInside(t *testing.T) {
	d := rectPointDistance(5, 5, 0, 0, 10, 10)
	if d != 0 {
		t.Errorf("point inside rect: got %f, want 0", d)
	}
}

func TestRectPointDistanceOnEdge(t *testing.T) {
	d := rectPointDistance(0, 5, 0, 0, 10, 10)
	if d != 0 {
		t.Errorf("point on edge: got %f, want 0", d)
	}
}

func TestRectPointDistanceOutside(t *testing.T) {
	tests := []struct {
		name         string
		px, py       float64
		x1, y1, x2, y2 float64
		want         float64
	}{
		{"right", 13, 5, 0, 0, 10, 10, 3},
		{"left", -4, 5, 0, 0, 10, 10, 4},
		{"above", 5, -3, 0, 0, 10, 10, 3},
		{"below", 5, 15, 0, 0, 10, 10, 5},
		{"corner", 13, 14, 0, 0, 10, 10, 5}, // 3-4-5 triangle
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rectPointDistance(tt.px, tt.py, tt.x1, tt.y1, tt.x2, tt.y2)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

// --- ovalPointDistance ---

func TestOvalPointDistanceInside(t *testing.T) {
	// Center of a circle radius 50
	d := ovalPointDistance(50, 50, 0, 0, 100, 100)
	if d != 0 {
		t.Errorf("center of circle: got %f, want 0", d)
	}
}

func TestOvalPointDistanceOutside(t *testing.T) {
	// Point 10px right of circle edge (circle centered at 50,50, radius 50)
	d := ovalPointDistance(110, 50, 0, 0, 100, 100)
	if d < 9 || d > 11 {
		t.Errorf("10px outside circle: got %f, want ~10", d)
	}
}

func TestOvalPointDistanceDegenerateZeroRadius(t *testing.T) {
	// Zero-size oval should return distance to center
	d := ovalPointDistance(3, 4, 0, 0, 0, 0)
	if math.Abs(d-5) > 0.001 {
		t.Errorf("degenerate oval: got %f, want 5", d)
	}
}

// --- segmentPointDistance ---

func TestSegmentPointDistanceOnSegment(t *testing.T) {
	d := segmentPointDistance(5, 0, 0, 0, 10, 0)
	if d != 0 {
		t.Errorf("point on segment: got %f, want 0", d)
	}
}

func TestSegmentPointDistancePerpendicular(t *testing.T) {
	d := segmentPointDistance(5, 3, 0, 0, 10, 0)
	if math.Abs(d-3) > 0.001 {
		t.Errorf("perpendicular distance: got %f, want 3", d)
	}
}

func TestSegmentPointDistanceBeyondEndpoint(t *testing.T) {
	// Point beyond end of segment — distance to endpoint
	d := segmentPointDistance(13, 4, 0, 0, 10, 0)
	want := 5.0 // 3-4-5 triangle from (10,0) to (13,4)
	if math.Abs(d-want) > 0.001 {
		t.Errorf("beyond endpoint: got %f, want %f", d, want)
	}
}

func TestSegmentPointDistanceZeroLength(t *testing.T) {
	d := segmentPointDistance(3, 4, 0, 0, 0, 0)
	if math.Abs(d-5) > 0.001 {
		t.Errorf("zero-length segment: got %f, want 5", d)
	}
}

// --- pointInPolygon ---

func TestPointInPolygonSquare(t *testing.T) {
	// Square: (0,0), (10,0), (10,10), (0,10)
	coords := []float64{0, 0, 10, 0, 10, 10, 0, 10}

	if !pointInPolygon(5, 5, coords) {
		t.Error("center should be inside square")
	}
	if pointInPolygon(15, 5, coords) {
		t.Error("point outside should not be inside square")
	}
}

func TestPointInPolygonTriangle(t *testing.T) {
	// Triangle: (0,0), (10,0), (5,10)
	coords := []float64{0, 0, 10, 0, 5, 10}

	if !pointInPolygon(5, 3, coords) {
		t.Error("point inside triangle should return true")
	}
	if pointInPolygon(0, 10, coords) {
		t.Error("point outside triangle should return false")
	}
}

func TestPointInPolygonConcave(t *testing.T) {
	// L-shaped polygon
	coords := []float64{0, 0, 10, 0, 10, 5, 5, 5, 5, 10, 0, 10}

	if !pointInPolygon(2, 2, coords) {
		t.Error("point in lower-left of L should be inside")
	}
	if pointInPolygon(8, 8, coords) {
		t.Error("point in upper-right cutout of L should be outside")
	}
}

// --- ItemBase tags ---

func TestItemBaseHasTag(t *testing.T) {
	b := ItemBase{Tags: []string{"foo", "bar"}}
	if !b.HasTag("foo") {
		t.Error("should have tag 'foo'")
	}
	if b.HasTag("baz") {
		t.Error("should not have tag 'baz'")
	}
}

func TestItemBaseAddTag(t *testing.T) {
	b := ItemBase{}
	b.AddTag("test")
	if !b.HasTag("test") {
		t.Error("tag should be added")
	}
	b.AddTag("test") // duplicate
	if len(b.Tags) != 1 {
		t.Errorf("duplicate tag added: len=%d", len(b.Tags))
	}
}

func TestItemBaseRemoveTag(t *testing.T) {
	b := ItemBase{Tags: []string{"a", "b", "c"}}
	b.RemoveTag("b")
	if b.HasTag("b") {
		t.Error("tag 'b' should be removed")
	}
	if len(b.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(b.Tags))
	}
	b.RemoveTag("nonexistent") // should not panic
}

// --- Spline generation ---

func TestGenerateBezierSplineTooFewPoints(t *testing.T) {
	// 1 point — returned as-is
	coords := []float64{5, 10}
	result := generateBezierSpline(coords, false, 12)
	if len(result) != 2 {
		t.Errorf("expected 2 values, got %d", len(result))
	}
}

func TestGenerateBezierSplineTwoPointsOpen(t *testing.T) {
	coords := []float64{0, 0, 10, 10}
	result := generateBezierSpline(coords, false, 12)
	// Should return copy of input
	if len(result) != 4 {
		t.Errorf("expected 4 values, got %d", len(result))
	}
	if result[0] != 0 || result[2] != 10 {
		t.Errorf("unexpected coords: %v", result)
	}
}

func TestGenerateBezierSplineOpenEndpoints(t *testing.T) {
	// 4 control points
	coords := []float64{0, 0, 10, 20, 30, 10, 40, 0}
	result := generateBezierSpline(coords, false, 4)

	// First point should be the input start
	if result[0] != 0 || result[1] != 0 {
		t.Errorf("first point should be (0,0), got (%f,%f)", result[0], result[1])
	}
	// Last point should be the input end
	lastX := result[len(result)-2]
	lastY := result[len(result)-1]
	if lastX != 40 || lastY != 0 {
		t.Errorf("last point should be (40,0), got (%f,%f)", lastX, lastY)
	}
	// Should have more points than input
	if len(result) <= len(coords) {
		t.Errorf("spline should generate more points: input=%d, output=%d", len(coords), len(result))
	}
}

func TestGenerateBezierSplineClosedWraps(t *testing.T) {
	// Triangle control points
	coords := []float64{0, 0, 10, 0, 5, 10}
	result := generateBezierSpline(coords, true, 4)

	// Closed spline should wrap: last point ≈ first point
	firstX, firstY := result[0], result[1]
	lastX, lastY := result[len(result)-2], result[len(result)-1]
	if math.Abs(firstX-lastX) > 0.001 || math.Abs(firstY-lastY) > 0.001 {
		t.Errorf("closed spline should wrap: first=(%f,%f), last=(%f,%f)", firstX, firstY, lastX, lastY)
	}
}

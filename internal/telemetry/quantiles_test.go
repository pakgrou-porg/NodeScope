package telemetry

import (
	"math"
	"testing"
)

func TestQuantileSketchCalculatesMergeableP95(t *testing.T) {
	left, err := NewQuantileSketch()
	if err != nil {
		t.Fatalf("create left sketch: %v", err)
	}
	right, err := NewQuantileSketch()
	if err != nil {
		t.Fatalf("create right sketch: %v", err)
	}

	for value := 1.0; value <= 100; value++ {
		target := left
		if value > 50 {
			target = right
		}
		if err := target.Add(value); err != nil {
			t.Fatalf("add %f: %v", value, err)
		}
	}
	if err := left.Merge(right); err != nil {
		t.Fatalf("merge: %v", err)
	}
	if left.Count() != 100 {
		t.Fatalf("expected merged count 100, got %f", left.Count())
	}
	p95, err := left.P95()
	if err != nil {
		t.Fatalf("p95: %v", err)
	}
	if relativeError := math.Abs(p95-95) / 95; relativeError > DefaultRelativeAccuracy*1.5 {
		t.Fatalf("expected p95 near 95, got %f (relative error %f)", p95, relativeError)
	}
}

func TestQuantileSketchRoundTrips(t *testing.T) {
	sketch, err := NewQuantileSketch()
	if err != nil {
		t.Fatalf("create sketch: %v", err)
	}
	for _, value := range []float64{2, 4, 8, 16, 32, 64, 128} {
		if err := sketch.Add(value); err != nil {
			t.Fatalf("add %f: %v", value, err)
		}
	}
	payload, err := sketch.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	decoded, err := UnmarshalQuantileSketch(sketch.Algorithm, sketch.FormatVersion, sketch.RelativeAccuracy, sketch.MaxBins, payload)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Count() != sketch.Count() {
		t.Fatalf("expected count %f, got %f", sketch.Count(), decoded.Count())
	}
	originalP95, err := sketch.P95()
	if err != nil {
		t.Fatalf("original p95: %v", err)
	}
	decodedP95, err := decoded.P95()
	if err != nil {
		t.Fatalf("decoded p95: %v", err)
	}
	if originalP95 != decodedP95 {
		t.Fatalf("expected p95 %f after round trip, got %f", originalP95, decodedP95)
	}
}

func TestQuantileSketchRejectsIncompatibleMerge(t *testing.T) {
	left, err := NewQuantileSketch()
	if err != nil {
		t.Fatalf("create left sketch: %v", err)
	}
	right, err := NewQuantileSketch()
	if err != nil {
		t.Fatalf("create right sketch: %v", err)
	}
	right.FormatVersion++
	if err := left.Merge(right); err == nil {
		t.Fatal("expected incompatible sketch merge to fail")
	}
}

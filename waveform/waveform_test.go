package waveform

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestSeismogramMergePreservesGaps(t *testing.T) {
	start := time.Unix(0, 0).UTC()
	first := SeismogramSegment{Data: []float64{1, 2}, StartTime: start, SampleRate: 2}
	second := SeismogramSegment{Data: []float64{4, 5}, StartTime: start.Add(2 * time.Second), SampleRate: 2}
	value, err := NewSeismogram(first, second)
	if err != nil {
		t.Fatal(err)
	}
	merged, err := value.Merge(DefaultMergeOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(merged.Segments))
	}
	if merged.IsContiguous(0.5) {
		t.Fatal("gap reported as contiguous")
	}
}

func TestSeismogramMergeGapPolicies(t *testing.T) {
	start := time.Unix(0, 0).UTC()
	value := Seismogram{Segments: []SeismogramSegment{
		{Data: []float64{1, 2}, StartTime: start, SampleRate: 1},
		{Data: []float64{5}, StartTime: start.Add(4 * time.Second), SampleRate: 1},
	}}
	merged, err := value.Merge(MergeOptions{GapPolicy: FillGapsWithNaN, ToleranceSamples: 0})
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.Segments) != 1 || len(merged.Segments[0].Data) != 5 {
		t.Fatalf("unexpected filled waveform: %#v", merged.Segments)
	}
	if !math.IsNaN(merged.Segments[0].Data[2]) || !math.IsNaN(merged.Segments[0].Data[3]) {
		t.Fatal("gap was not filled with NaN")
	}
}

func TestSeismogramRejectsConflictingOverlap(t *testing.T) {
	start := time.Unix(0, 0).UTC()
	value := Seismogram{Segments: []SeismogramSegment{
		{Data: []float64{1, 2, 3}, StartTime: start, SampleRate: 1},
		{Data: []float64{9, 4}, StartTime: start.Add(2 * time.Second), SampleRate: 1},
	}}
	_, err := value.Merge(MergeOptions{GapPolicy: PreserveGaps, OverlapPolicy: RejectOverlaps, ToleranceSamples: 0})
	if !errors.Is(err, ErrOverlappingData) {
		t.Fatalf("error = %v, want conflicting overlap", err)
	}
}

func TestSegmentTimeSemantics(t *testing.T) {
	segment := SeismogramSegment{Data: []float64{1, 2, 3}, StartTime: time.Unix(0, 0), SampleRate: 2}
	if got, want := segment.EndTime(), time.Unix(1, 0); !got.Equal(want) {
		t.Fatalf("end = %s, want %s", got, want)
	}
	if got, want := segment.EndExclusive(), time.Unix(0, 1500*int64(time.Millisecond)); !got.Equal(want) {
		t.Fatalf("end exclusive = %s, want %s", got, want)
	}
}

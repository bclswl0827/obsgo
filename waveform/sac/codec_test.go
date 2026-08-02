package sac

import (
	"math"
	"testing"
	"time"

	"github.com/bclswl0827/obsgo/waveform"
)

func TestRoundTrip(t *testing.T) {
	start := time.Date(2026, 8, 1, 1, 2, 3, 4000000, time.UTC)
	stream := waveform.NewStream(waveform.NewTrace("XX", "TEST", "00", "BHZ", waveform.Seismogram{
		Unit: "m/s", SourceID: "XX.TEST.00.BHZ", Segments: []waveform.SeismogramSegment{{
			Data: []float64{1.25, -2, 3}, StartTime: start, SampleRate: 20, Unit: "m/s", SourceID: "XX.TEST.00.BHZ",
		}},
	}))
	path := t.TempDir() + "/test.sac"
	if err := Write(path, stream); err != nil {
		t.Fatal(err)
	}
	decoded, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	segment := decoded.Traces[0].Segments[0]
	if len(segment.Data) != 3 || segment.Data[0] != float64(float32(1.25)) {
		t.Fatalf("data = %#v", segment.Data)
	}
	if math.Abs(segment.SampleRate-20) > 1e-5 {
		t.Fatalf("sample rate = %g", segment.SampleRate)
	}
	if segment.Unit != "m/s" {
		t.Fatalf("unit = %q", segment.Unit)
	}
	if !segment.StartTime.Equal(start) {
		t.Fatalf("start = %s, want %s", segment.StartTime, start)
	}
}

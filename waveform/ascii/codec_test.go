package ascii

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/bclswl0827/obsgo/waveform"
)

func TestRoundTrip(t *testing.T) {
	start := time.Date(2026, 8, 1, 1, 2, 3, 4000000, time.UTC)
	stream := waveform.NewStream(waveform.NewTrace("XX", "TEST", "00", "BHZ", waveform.Seismogram{
		Unit: "m/s", SourceID: "XX.TEST.00.BHZ", Segments: []waveform.SeismogramSegment{
			{Data: []float64{1.25, -2, 3}, StartTime: start, SampleRate: 20, Unit: "m/s", SourceID: "XX.TEST.00.BHZ"},
			{Data: []float64{4}, StartTime: start.Add(time.Second), SampleRate: 20, Unit: "m/s", SourceID: "XX.TEST.00.BHZ"},
		},
	}))
	var encoded bytes.Buffer
	if err := WriteWriter(&encoded, stream); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadReader(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Traces[0].Segments, stream.Traces[0].Segments) {
		t.Fatalf("segments = %#v, want %#v", decoded.Traces[0].Segments, stream.Traces[0].Segments)
	}
	if decoded.Traces[0].SourceID() != stream.Traces[0].SourceID() {
		t.Fatalf("source ID = %q", decoded.Traces[0].SourceID())
	}
}

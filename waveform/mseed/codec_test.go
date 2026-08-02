package mseed

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/bclswl0827/mseedio"
	"github.com/bclswl0827/obsgo/waveform"
)

func TestRoundTripAndRecordMerge(t *testing.T) {
	start := time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC)
	stream := waveform.NewStream(waveform.NewTrace("XX", "TEST", "00", "BHZ", waveform.Seismogram{
		SourceID: "XX.TEST.00.BHZ", Segments: []waveform.SeismogramSegment{{
			Data: []float64{1, 2, 3, 4}, StartTime: start, SampleRate: 2, SourceID: "XX.TEST.00.BHZ",
		}},
	}))
	var encoded bytes.Buffer
	if err := WriteWriterWithOptions(&encoded, stream, WriteOptions{Encoding: mseedio.INT32, BitOrder: mseedio.MSBFIRST, SamplesPerRecord: 2}); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadReader(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Len() != 1 || len(decoded.Traces[0].Segments) != 1 {
		t.Fatalf("decoded stream = %#v", decoded)
	}
	if len(decoded.Traces[0].Segments[0].Data) != 4 {
		t.Fatalf("decoded samples = %d", len(decoded.Traces[0].Segments[0].Data))
	}
}

func TestReadRepositoryFixture(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "_data", "2026.211.16.53.54.0541.SHAKE.AS.00.EHZ.D.mseed")
	stream, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if stream.Len() != 1 {
		t.Fatalf("traces = %d, want 1", stream.Len())
	}
	trace := stream.Traces[0]
	if trace.Network != "AS" || trace.Station != "SHAKE" || trace.Location != "00" || trace.Channel != "EHZ" {
		t.Fatalf("trace ID = %s", trace.SourceID())
	}
	if len(trace.Segments) == 0 || len(trace.Segments[0].Data) == 0 {
		t.Fatalf("segments = %#v", trace.Segments)
	}
}

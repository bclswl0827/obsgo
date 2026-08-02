package waveform

import (
	"math"
	"testing"
	"time"

	"github.com/bclswl0827/obsgo/inventory"
	"github.com/bclswl0827/obsgo/signal"
)

func TestCalculusUpdatesUnitsAndTiming(t *testing.T) {
	start := time.Unix(100, 0).UTC()
	seismogram := Seismogram{
		Segments: []SeismogramSegment{{
			Data:       []float64{1, 3, 6},
			StartTime:  start,
			SampleRate: 2,
			Unit:       "m",
			SourceID:   "XX.TEST.00.BHZ",
		}},
		Unit:       "m",
		SourceID:   "XX.TEST.00.BHZ",
		SampleRate: 2,
		StartTime:  start,
	}
	differentiated, err := seismogram.Differentiate()
	if err != nil {
		t.Fatal(err)
	}
	if differentiated.Unit != "m/s" || differentiated.Segments[0].Unit != "m/s" {
		t.Fatalf("differentiate units = (%q, %q), want m/s", differentiated.Unit, differentiated.Segments[0].Unit)
	}
	wantStart := start.Add(250 * time.Millisecond)
	if !differentiated.StartTime.Equal(wantStart) || !differentiated.Segments[0].StartTime.Equal(wantStart) {
		t.Fatalf("differentiate starts = (%s, %s), want %s", differentiated.StartTime, differentiated.Segments[0].StartTime, wantStart)
	}

	gradient, err := seismogram.DifferentiateGradient()
	if err != nil {
		t.Fatal(err)
	}
	if gradient.Unit != "m/s" || !gradient.Segments[0].StartTime.Equal(start) {
		t.Fatalf("gradient metadata = (%q, %s), want (m/s, %s)", gradient.Unit, gradient.Segments[0].StartTime, start)
	}
}

func TestIntegrateCarriesOnlyAcrossContinuousSegments(t *testing.T) {
	start := time.Unix(0, 0).UTC()
	continuous := Seismogram{
		Segments: []SeismogramSegment{
			{Data: []float64{1, 1}, StartTime: start, SampleRate: 2, Unit: "m/s^2"},
			{Data: []float64{2}, StartTime: start.Add(time.Second), SampleRate: 2, Unit: "m/s^2"},
		},
		Unit: "m/s^2", SampleRate: 2, StartTime: start,
	}
	integrated, err := continuous.Integrate(3)
	if err != nil {
		t.Fatal(err)
	}
	assertWaveformSamples(t, "first continuous integral", integrated.Segments[0].Data, []float64{3, 3.5, 4})
	assertWaveformSamples(t, "second continuous integral", integrated.Segments[1].Data, []float64{4, 5})
	wantStart := start.Add(-250 * time.Millisecond)
	if !integrated.StartTime.Equal(wantStart) || !integrated.Segments[0].StartTime.Equal(wantStart) {
		t.Fatalf("integrate starts = (%s, %s), want %s", integrated.StartTime, integrated.Segments[0].StartTime, wantStart)
	}
	if integrated.Unit != "m/s" || integrated.Segments[0].Unit != "m/s" || integrated.Segments[1].Unit != "m/s" {
		t.Fatalf("integrate units = (%q, %q, %q), want m/s", integrated.Unit, integrated.Segments[0].Unit, integrated.Segments[1].Unit)
	}

	gapped := continuous.Clone()
	gapped.Segments[1].StartTime = start.Add(2 * time.Second)
	integrated, err = gapped.Integrate(3)
	if err != nil {
		t.Fatal(err)
	}
	assertWaveformSamples(t, "integral after gap", integrated.Segments[1].Data, []float64{3, 4})
}

func TestTraceSignalWrappersPreserveMetadata(t *testing.T) {
	start := time.Unix(0, 0).UTC()
	seismogram := Seismogram{
		Segments: []SeismogramSegment{{
			Data: []float64{1, -0.5, 0.25, 2, -1, 0.125}, StartTime: start,
			SampleRate: 20, Unit: "m/s", SourceID: "XX.TEST.00.BHZ",
		}},
		Unit: "m/s", SourceID: "XX.TEST.00.BHZ", SampleRate: 20, StartTime: start,
	}
	trace := NewTrace("XX", "TEST", "00", "BHZ", seismogram)

	envelope := trace.ObsPyEnvelope()
	if envelope.Network != trace.Network || envelope.Station != trace.Station || envelope.Channel != trace.Channel {
		t.Fatal("ObsPy envelope changed Trace channel metadata")
	}
	if len(envelope.Segments[0].Data) != len(trace.Segments[0].Data) || envelope.Unit != "m/s" {
		t.Fatalf("ObsPy envelope metadata = (length %d, unit %q)", len(envelope.Segments[0].Data), envelope.Unit)
	}

	centered := trace.CenteredHilbert()
	if got, want := len(centered.Segments[0].Data), len(trace.Segments[0].Data)+20; got != want {
		t.Fatalf("centered Hilbert length = %d, want %d", got, want)
	}

	resampled, err := trace.Resample(40, signal.LinearInterpolation, signal.InterpolationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if resampled.SampleRate != 40 || resampled.Segments[0].SampleRate != 40 || resampled.SourceID() != trace.SourceID() {
		t.Fatalf("resample metadata = (%g, %g, %q)", resampled.SampleRate, resampled.Segments[0].SampleRate, resampled.SourceID())
	}
}

func TestTraceRotationUpdatesStandardComponentIDs(t *testing.T) {
	start := time.Unix(0, 0).UTC()
	makeTrace := func(component string, data []float64) Trace {
		sourceID := "XX.TEST.00.BH" + component
		seismogram := Seismogram{
			Segments: []SeismogramSegment{{Data: data, StartTime: start, SampleRate: 20, Unit: "m/s", SourceID: sourceID}},
			Unit:     "m/s", SourceID: sourceID, SampleRate: 20, StartTime: start,
		}
		return NewTrace("XX", "TEST", "00", "BH"+component, seismogram)
	}
	vertical := makeTrace("Z", []float64{1, 2})
	north := makeTrace("N", []float64{3, 4})
	east := makeTrace("E", []float64{5, 6})
	_, radial, transverse, err := RotateTraceZNEToZRT(vertical, north, east, 30)
	if err != nil {
		t.Fatal(err)
	}
	if radial.Channel != "BHR" || radial.Seismogram.SourceID != "XX.TEST.00.BHR" || radial.Segments[0].SourceID != "XX.TEST.00.BHR" {
		t.Fatalf("radial IDs = (%q, %q, %q)", radial.Channel, radial.Seismogram.SourceID, radial.Segments[0].SourceID)
	}
	if transverse.Channel != "BHT" || transverse.Seismogram.SourceID != "XX.TEST.00.BHT" {
		t.Fatalf("transverse IDs = (%q, %q)", transverse.Channel, transverse.Seismogram.SourceID)
	}
}

func TestSpectralAndSimulationUnits(t *testing.T) {
	segment := SeismogramSegment{
		Data: []float64{1, 2, 3, 4}, StartTime: time.Unix(0, 0).UTC(),
		SampleRate: 20, Unit: "m/s", SourceID: "XX.TEST.00.BHZ",
	}
	seismogram := Seismogram{Segments: []SeismogramSegment{segment}, Unit: segment.Unit, SourceID: segment.SourceID, SampleRate: segment.SampleRate}
	psd, err := seismogram.PowerSpectralDensity(4, signal.HannWindow, signal.SpectrumOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if psd[0].Unit != "(m/s)^2/Hz" {
		t.Fatalf("PSD unit = %q, want (m/s)^2/Hz", psd[0].Unit)
	}
	legacy, err := seismogram.ObsPySpectrum(4, signal.HannWindow)
	if err != nil {
		t.Fatal(err)
	}
	if legacy[0].Unit != "(m/s)^2" {
		t.Fatalf("ObsPy spectrum unit = %q, want (m/s)^2", legacy[0].Unit)
	}

	response := &inventory.Response{Sensitivity: &inventory.InstrumentSensitivity{Output: inventory.Units{Name: "count"}}}
	if got := simulatedResponseUnit(response); got != "count" {
		t.Fatalf("simulated response unit = %q, want count", got)
	}
}

func assertWaveformSamples(t *testing.T, name string, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", name, len(got), len(want))
	}
	for index := range got {
		if math.Abs(got[index]-want[index]) > 1e-12 {
			t.Fatalf("%s[%d] = %.17g, want %.17g", name, index, got[index], want[index])
		}
	}
}

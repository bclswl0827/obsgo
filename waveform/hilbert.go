package waveform

import "github.com/bclswl0827/obsgo/signal"

// ObsPyHilbert applies the exact-length FFT Hilbert transform independently
// to every contiguous segment.
func (seismogram Seismogram) ObsPyHilbert() Seismogram {
	result := seismogram.Clone()
	for index := range result.Segments {
		result.Segments[index].Data = signal.ObsPyHilbert(result.Segments[index].Data)
	}
	return result
}

// ObsPyEnvelope returns the FFT analytic-signal envelope for every segment.
func (seismogram Seismogram) ObsPyEnvelope() Seismogram {
	result := seismogram.Clone()
	for index := range result.Segments {
		result.Segments[index].Data = signal.ObsPyEnvelope(result.Segments[index].Data)
	}
	return result
}

// CenteredHilbert applies the OregonDSP centered FIR and returns its full
// convolution for each segment.
func (seismogram Seismogram) CenteredHilbert() Seismogram {
	result := seismogram.Clone()
	for index := range result.Segments {
		result.Segments[index].Data = signal.CenteredHilbert(result.Segments[index].Data)
	}
	return result
}

// CenteredEnvelope returns the OregonDSP centered-FIR envelope for every
// segment.
func (seismogram Seismogram) CenteredEnvelope() Seismogram {
	result := seismogram.Clone()
	for index := range result.Segments {
		result.Segments[index].Data = signal.CenteredEnvelope(result.Segments[index].Data)
	}
	return result
}

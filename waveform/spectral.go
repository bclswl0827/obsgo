package waveform

import (
	"time"

	"github.com/bclswl0827/obsgo/signal"
)

// SpectrumSegment associates an array-level spectrum with its waveform
// segment and source metadata.
type SpectrumSegment struct {
	StartTime time.Time
	SourceID  string
	Unit      string
	Value     *signal.SpectrumResult
}

// PSDWaveformSegment associates a PSD estimate with its source segment.
type PSDWaveformSegment struct {
	StartTime time.Time
	SourceID  string
	Unit      string
	Value     *signal.PSDResult
}

// TwoSidedSpectrumSegment associates an ObsPy two-sided spectrum with a
// waveform segment.
type TwoSidedSpectrumSegment struct {
	StartTime time.Time
	SourceID  string
	Unit      string
	Value     *signal.TwoSidedSpectrumResult
}

// ComputeSpectrum computes one spectrum per contiguous segment.
func (seismogram Seismogram) ComputeSpectrum(nfft int, window signal.SpectralWindow) ([]SpectrumSegment, error) {
	result := make([]SpectrumSegment, 0, len(seismogram.Segments))
	for _, segment := range seismogram.Segments {
		value, err := signal.ComputeSpectrum(segment.Data, segment.SampleRate, nfft, window)
		if err != nil {
			return nil, err
		}
		result = append(result, SpectrumSegment{StartTime: segment.StartTime, SourceID: segment.SourceID, Unit: segment.Unit, Value: value})
	}
	return result, nil
}

// PowerSpectralDensity computes one PSD per contiguous segment.
func (seismogram Seismogram) PowerSpectralDensity(nfft int, window signal.SpectralWindow, options signal.SpectrumOptions) ([]PSDWaveformSegment, error) {
	result := make([]PSDWaveformSegment, 0, len(seismogram.Segments))
	for _, segment := range seismogram.Segments {
		value, err := signal.PowerSpectralDensityWithOptions(segment.Data, segment.SampleRate, nfft, window, options)
		if err != nil {
			return nil, err
		}
		result = append(result, PSDWaveformSegment{StartTime: segment.StartTime, SourceID: segment.SourceID, Unit: powerDensityUnit(segment.Unit), Value: value})
	}
	return result, nil
}

// Welch computes one averaged PSD per contiguous segment. Welch windows do
// not cross gaps because each segment is passed separately to signal.Welch.
func (seismogram Seismogram) Welch(options signal.WelchOptions) ([]PSDWaveformSegment, error) {
	result := make([]PSDWaveformSegment, 0, len(seismogram.Segments))
	for _, segment := range seismogram.Segments {
		value, err := signal.Welch(segment.Data, segment.SampleRate, options)
		if err != nil {
			return nil, err
		}
		result = append(result, PSDWaveformSegment{StartTime: segment.StartTime, SourceID: segment.SourceID, Unit: powerDensityUnit(segment.Unit), Value: value})
	}
	return result, nil
}

// ObsPySpectrum computes the unshifted two-sided ObsPy spectrum per segment.
func (seismogram Seismogram) ObsPySpectrum(nfft int, window signal.SpectralWindow) ([]TwoSidedSpectrumSegment, error) {
	result := make([]TwoSidedSpectrumSegment, 0, len(seismogram.Segments))
	for _, segment := range seismogram.Segments {
		value, err := signal.ObsPySpectrum(segment.Data, segment.SampleRate, nfft, window)
		if err != nil {
			return nil, err
		}
		result = append(result, TwoSidedSpectrumSegment{StartTime: segment.StartTime, SourceID: segment.SourceID, Unit: powerUnit(segment.Unit), Value: value})
	}
	return result, nil
}

// ObsPyWelch computes the unshifted two-sided ObsPy Welch spectrum per
// segment.
func (seismogram Seismogram) ObsPyWelch(segmentLength, nfft int, overlapFraction float64, window signal.SpectralWindow) ([]TwoSidedSpectrumSegment, error) {
	result := make([]TwoSidedSpectrumSegment, 0, len(seismogram.Segments))
	for _, segment := range seismogram.Segments {
		value, err := signal.ObsPyWelch(segment.Data, segment.SampleRate, segmentLength, nfft, overlapFraction, window)
		if err != nil {
			return nil, err
		}
		result = append(result, TwoSidedSpectrumSegment{StartTime: segment.StartTime, SourceID: segment.SourceID, Unit: powerUnit(segment.Unit), Value: value})
	}
	return result, nil
}

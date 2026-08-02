package signal

import (
	"fmt"
	"math"
	"math/cmplx"
)

// SpectralWindow selects the deterministic window used by spectral methods.
type SpectralWindow int

const (
	DefaultSpectralWindow SpectralWindow = iota
	RectangularWindow
	HannWindow
	HammingWindow
	WelchWindow
)

// SpectrumResult contains a one-sided spectrum from DC through Nyquist.
type SpectrumResult struct {
	Frequencies []float64
	Values      []complex128
	Amplitude   []float64
	Power       []float64
	SampleRate  float64
	NFFT        int
}

// SpectrumOptions controls preprocessing for ComputeSpectrum. The default
// SeisPlotJS-style spectrum demeans the input; ObsPySpectrum does not.
type SpectrumOptions struct {
	Demean bool
}

// ComputeSpectrum calculates a one-sided, windowed complex spectrum. The
// complex values use the unnormalized FFT convention; Amplitude is normalized
// by the coherent window gain and doubled away from DC/Nyquist.
func ComputeSpectrum(data []float64, sampleRate float64, nfft int, window SpectralWindow) (*SpectrumResult, error) {
	return ComputeSpectrumWithOptions(data, sampleRate, nfft, window, SpectrumOptions{Demean: true})
}

// ComputeSpectrumWithOptions calculates a one-sided amplitude spectrum with
// explicit preprocessing semantics.
func ComputeSpectrumWithOptions(data []float64, sampleRate float64, nfft int, window SpectralWindow, options SpectrumOptions) (*SpectrumResult, error) {
	if !finitePositive(sampleRate) {
		return nil, fmt.Errorf("sample rate must be finite and positive")
	}
	if len(data) == 0 {
		return &SpectrumResult{Frequencies: []float64{}, Values: []complex128{}, Amplitude: []float64{}, Power: []float64{}, SampleRate: sampleRate}, nil
	}
	if nfft <= 0 {
		nfft, _ = FindPowerTwo(len(data))
	}
	if nfft < len(data) {
		return nil, fmt.Errorf("nfft %d is shorter than data length %d", nfft, len(data))
	}
	weights, err := spectralWindow(window, len(data))
	if err != nil {
		return nil, err
	}
	values := make([]complex128, nfft)
	mean := 0.0
	if options.Demean {
		mean = Mean(data)
	}
	for index, value := range data {
		values[index] = complex((value-mean)*weights[index], 0)
	}
	fftExactInPlace(values, false)
	oneSided := nfft/2 + 1
	result := &SpectrumResult{
		Frequencies: make([]float64, oneSided),
		Values:      append([]complex128(nil), values[:oneSided]...),
		Amplitude:   make([]float64, oneSided),
		Power:       make([]float64, oneSided),
		SampleRate:  sampleRate,
		NFFT:        nfft,
	}
	coherentGain := 0.0
	for _, weight := range weights {
		coherentGain += weight
	}
	if coherentGain == 0 {
		return nil, fmt.Errorf("spectral window has zero coherent gain")
	}
	for index, value := range result.Values {
		result.Frequencies[index] = float64(index) * sampleRate / float64(nfft)
		scale := 1.0 / coherentGain
		if index != 0 && !(nfft%2 == 0 && index == nfft/2) {
			scale *= 2
		}
		result.Amplitude[index] = cmplx.Abs(value) * scale
		result.Power[index] = result.Amplitude[index] * result.Amplitude[index]
	}
	return result, nil
}

// PSDResult is a one-sided power spectral density estimate in units squared
// per hertz.
type PSDResult struct {
	Frequencies []float64
	Power       []float64
	SampleRate  float64
	NFFT        int
}

// PowerSpectralDensity calculates a single-segment periodogram using the
// density normalization sum(window^2)*sampleRate. The default does not
// detrend; callers can request demeaning through
// PowerSpectralDensityWithOptions.
func PowerSpectralDensity(data []float64, sampleRate float64, nfft int, window SpectralWindow) (*PSDResult, error) {
	return PowerSpectralDensityWithOptions(data, sampleRate, nfft, window, SpectrumOptions{})
}

// PowerSpectralDensityWithOptions calculates a one-sided density estimate.
func PowerSpectralDensityWithOptions(data []float64, sampleRate float64, nfft int, window SpectralWindow, options SpectrumOptions) (*PSDResult, error) {
	if !finitePositive(sampleRate) {
		return nil, fmt.Errorf("sample rate must be finite and positive")
	}
	if len(data) == 0 {
		return &PSDResult{Frequencies: []float64{}, Power: []float64{}, SampleRate: sampleRate}, nil
	}
	if nfft <= 0 {
		nfft, _ = FindPowerTwo(len(data))
	}
	if nfft < len(data) {
		return nil, fmt.Errorf("nfft %d is shorter than data length %d", nfft, len(data))
	}
	values, err := periodogram(data, nfft, sampleRate, window, options.Demean)
	if err != nil {
		return nil, err
	}
	return &PSDResult{Frequencies: frequencies(sampleRate, nfft), Power: values, SampleRate: sampleRate, NFFT: nfft}, nil
}

// WelchOptions controls averaged periodogram estimation.
type WelchOptions struct {
	SegmentLength int
	// OverlapFraction is nil for SciPy/ObsPy's 50% default. A non-nil zero
	// explicitly disables overlap.
	OverlapFraction *float64
	NFFT            int
	Window          SpectralWindow
	Detrend         SpectralDetrend
}

// SpectralDetrend selects segment preprocessing for Welch estimation.
type SpectralDetrend int

const (
	DetrendConstant SpectralDetrend = iota
	DetrendNone
)

// DefaultWelchOptions returns the SciPy/ObsPy convention: 256-sample Hann
// windows, 50% overlap, constant detrending, and nfft equal to nperseg.
func DefaultWelchOptions() WelchOptions {
	overlap := 0.5
	return WelchOptions{
		OverlapFraction: &overlap,
		Window:          HannWindow,
		Detrend:         DetrendConstant,
	}
}

// Welch calculates an averaged one-sided PSD using complete overlapping
// segments. OverlapFraction follows ObsPy's 0..1 fractional convention and no
// incomplete tail segment is zero-padded.
func Welch(data []float64, sampleRate float64, options WelchOptions) (*PSDResult, error) {
	if !finitePositive(sampleRate) {
		return nil, fmt.Errorf("sample rate must be finite and positive")
	}
	if len(data) == 0 {
		return &PSDResult{Frequencies: []float64{}, Power: []float64{}, SampleRate: sampleRate}, nil
	}
	segmentLength := options.SegmentLength
	if segmentLength <= 0 {
		segmentLength = len(data)
		if segmentLength > 256 {
			segmentLength = 256
		}
	}
	if segmentLength <= 0 {
		return nil, fmt.Errorf("segment length must be positive")
	}
	overlapFraction := 0.5
	if options.OverlapFraction != nil {
		overlapFraction = *options.OverlapFraction
	}
	if math.IsNaN(overlapFraction) || math.IsInf(overlapFraction, 0) || overlapFraction < 0 || overlapFraction >= 1 {
		return nil, fmt.Errorf("overlap fraction must be in [0, 1)")
	}
	nfft := options.NFFT
	if nfft <= 0 {
		nfft = segmentLength
	}
	if nfft < segmentLength {
		return nil, fmt.Errorf("nfft %d is shorter than segment length %d", nfft, segmentLength)
	}
	noOverlap := int(float64(segmentLength) * overlapFraction)
	step := segmentLength - noOverlap
	if step < 1 {
		return nil, fmt.Errorf("overlap fraction leaves no forward progress")
	}
	if len(data) < segmentLength {
		return nil, fmt.Errorf("data length %d is shorter than segment length %d", len(data), segmentLength)
	}
	segments := 1 + (len(data)-segmentLength)/step
	average := make([]float64, nfft/2+1)
	for segmentIndex := 0; segmentIndex < segments; segmentIndex++ {
		start := segmentIndex * step
		end := start + segmentLength
		segment := data[start:end]
		values, err := periodogram(segment, nfft, sampleRate, options.Window, options.Detrend == DetrendConstant)
		if err != nil {
			return nil, err
		}
		for index := range average {
			average[index] += values[index]
		}
	}
	for index := range average {
		average[index] /= float64(segments)
	}
	return &PSDResult{Frequencies: frequencies(sampleRate, nfft), Power: average, SampleRate: sampleRate, NFFT: nfft}, nil
}

// TwoSidedSpectrumResult is the unshifted full-frequency power spectrum used
// by ObsPy's freqattributes.spectrum and freqattributes.welch routines.
type TwoSidedSpectrumResult struct {
	Frequencies []float64
	Power       []float64
	SampleRate  float64
	NFFT        int
}

// ObsPySpectrum reproduces the legacy ObsPy modified-periodogram convention:
// no implicit demeaning, power divided by sum(window^2), and DC copied from
// the first positive-frequency bin.
func ObsPySpectrum(data []float64, sampleRate float64, nfft int, window SpectralWindow) (*TwoSidedSpectrumResult, error) {
	if sampleRate <= 0 || math.IsNaN(sampleRate) || math.IsInf(sampleRate, 0) {
		return nil, fmt.Errorf("sample rate must be finite and positive")
	}
	if len(data) == 0 {
		return &TwoSidedSpectrumResult{Frequencies: []float64{}, Power: []float64{}, SampleRate: sampleRate}, nil
	}
	if nfft <= 0 {
		nfft, _ = FindPowerTwo(len(data))
	}
	if nfft < len(data) {
		return nil, fmt.Errorf("nfft %d is shorter than data length %d", nfft, len(data))
	}
	weights, err := obsPySpectralWindow(window, len(data))
	if err != nil {
		return nil, err
	}
	values := make([]complex128, nfft)
	windowPower := 0.0
	for index, value := range data {
		values[index] = complex(value*weights[index], 0)
		windowPower += weights[index] * weights[index]
	}
	if windowPower == 0 {
		return nil, fmt.Errorf("spectral window has zero power")
	}
	fftExactInPlace(values, false)
	power := make([]float64, nfft)
	for index, value := range values {
		power[index] = cmplx.Abs(value) * cmplx.Abs(value) / windowPower
	}
	if len(power) > 1 {
		power[0] = power[1]
	}
	return &TwoSidedSpectrumResult{Frequencies: fullFrequencies(sampleRate, nfft), Power: power, SampleRate: sampleRate, NFFT: nfft}, nil
}

// ObsPyWelch averages ObsPySpectrum over complete windows using a fractional
// overlap. It returns the full, unshifted spectrum.
func ObsPyWelch(data []float64, sampleRate float64, segmentLength, nfft int, overlapFraction float64, window SpectralWindow) (*TwoSidedSpectrumResult, error) {
	if !finitePositive(sampleRate) {
		return nil, fmt.Errorf("sample rate must be finite and positive")
	}
	if len(data) == 0 {
		return &TwoSidedSpectrumResult{Frequencies: []float64{}, Power: []float64{}, SampleRate: sampleRate}, nil
	}
	if segmentLength <= 0 {
		segmentLength = len(data)
	}
	if overlapFraction < 0 || overlapFraction >= 1 || math.IsNaN(overlapFraction) || math.IsInf(overlapFraction, 0) {
		return nil, fmt.Errorf("overlap fraction must be in [0, 1)")
	}
	if len(data) < segmentLength {
		return nil, fmt.Errorf("data length %d is shorter than segment length %d", len(data), segmentLength)
	}
	noOverlap := int(float64(segmentLength) * overlapFraction)
	step := segmentLength - noOverlap
	if step < 1 {
		return nil, fmt.Errorf("overlap fraction leaves no forward progress")
	}
	segments := 1 + (len(data)-segmentLength)/step
	var result *TwoSidedSpectrumResult
	for index := 0; index < segments; index++ {
		part, err := ObsPySpectrum(data[index*step:index*step+segmentLength], sampleRate, nfft, window)
		if err != nil {
			return nil, err
		}
		if result == nil {
			result = part
			result.Power = make([]float64, len(part.Power))
		}
		for frequency := range result.Power {
			result.Power[frequency] += part.Power[frequency]
		}
	}
	for frequency := range result.Power {
		result.Power[frequency] /= float64(segments)
	}
	return result, nil
}

func periodogram(data []float64, nfft int, sampleRate float64, window SpectralWindow, demean bool) ([]float64, error) {
	// scipy.signal.welch uses DFT-even (periodic) windows. The separate
	// ObsPySpectrum API retains ObsPy's legacy symmetric-window convention.
	weights, err := spectralWindow(window, len(data))
	if err != nil {
		return nil, err
	}
	values := make([]complex128, nfft)
	mean := 0.0
	if demean {
		mean = Mean(data)
	}
	windowPower := 0.0
	for index, value := range data {
		values[index] = complex((value-mean)*weights[index], 0)
		windowPower += weights[index] * weights[index]
	}
	if windowPower == 0 {
		return nil, fmt.Errorf("spectral window has zero power")
	}
	fftExactInPlace(values, false)
	power := make([]float64, nfft/2+1)
	for index := range power {
		power[index] = cmplx.Abs(values[index]) * cmplx.Abs(values[index]) / (sampleRate * windowPower)
		if index != 0 && !(nfft%2 == 0 && index == nfft/2) {
			power[index] *= 2
		}
	}
	return power, nil
}

func frequencies(sampleRate float64, nfft int) []float64 {
	result := make([]float64, nfft/2+1)
	for index := range result {
		result[index] = float64(index) * sampleRate / float64(nfft)
	}
	return result
}

func fullFrequencies(sampleRate float64, nfft int) []float64 {
	result := make([]float64, nfft)
	for index := range result {
		result[index] = float64(index) * sampleRate / float64(nfft)
	}
	return result
}

package signal

import (
	"fmt"
	"math"
)

func NewButterworth(order int, passband Passband, lowFrequency, highFrequency, samplePeriod float64) (*IIRFilter, error) {
	if order <= 0 {
		return nil, fmt.Errorf("filter order must be positive")
	}
	return newOregonIIRFilter(
		newOregonButterworth(order),
		passband,
		lowFrequency,
		highFrequency,
		samplePeriod,
	)
}

// CreateButterworth is the SeisPlotJS-compatible constructor name.
func CreateButterworth(order int, passband Passband, lowFrequency, highFrequency, samplePeriod float64) (*IIRFilter, error) {
	return NewButterworth(order, passband, lowFrequency, highFrequency, samplePeriod)
}

func Lowpass(data []float64, cutoffFrequency, sampleRate float64, order int, zeroPhase bool) ([]float64, error) {
	period, err := samplePeriodForRate(sampleRate)
	if err != nil {
		return nil, err
	}
	filter, err := NewButterworth(order, LowPass, 0, cutoffFrequency, period)
	if err != nil {
		return nil, err
	}
	if zeroPhase {
		return filter.ApplyZeroPhase(data), nil
	}
	return filter.Apply(data), nil
}

func Highpass(data []float64, cutoffFrequency, sampleRate float64, order int, zeroPhase bool) ([]float64, error) {
	period, err := samplePeriodForRate(sampleRate)
	if err != nil {
		return nil, err
	}
	filter, err := NewButterworth(order, HighPass, cutoffFrequency, 0, period)
	if err != nil {
		return nil, err
	}
	if zeroPhase {
		return filter.ApplyZeroPhase(data), nil
	}
	return filter.Apply(data), nil
}

func BandpassFilter(data []float64, lowFrequency, highFrequency, sampleRate float64, order int, zeroPhase bool) ([]float64, error) {
	period, err := samplePeriodForRate(sampleRate)
	if err != nil {
		return nil, err
	}
	filter, err := NewButterworth(order, BandPass, lowFrequency, highFrequency, period)
	if err != nil {
		return nil, err
	}
	if zeroPhase {
		return filter.ApplyZeroPhase(data), nil
	}
	return filter.Apply(data), nil
}

func Bandpass(data []float64, lowFrequency, highFrequency, sampleRate float64, order int, zeroPhase bool) ([]float64, error) {
	return BandpassFilter(data, lowFrequency, highFrequency, sampleRate, order, zeroPhase)
}

// ObsPyLowpass applies a Float64 Butterworth low-pass using ObsPy's
// unpadded sosfilt forward/reverse zero-phase convention.
func ObsPyLowpass(data []float64, cutoffFrequency, sampleRate float64, order int, zeroPhase bool) ([]float64, error) {
	return obsPyButterworth(data, LowPass, 0, cutoffFrequency, sampleRate, order, zeroPhase)
}

// ObsPyHighpass applies a Float64 Butterworth high-pass.
func ObsPyHighpass(data []float64, cutoffFrequency, sampleRate float64, order int, zeroPhase bool) ([]float64, error) {
	return obsPyButterworth(data, HighPass, cutoffFrequency, 0, sampleRate, order, zeroPhase)
}

// ObsPyBandpass applies a Float64 Butterworth band-pass.
func ObsPyBandpass(data []float64, lowFrequency, highFrequency, sampleRate float64, order int, zeroPhase bool) ([]float64, error) {
	return obsPyButterworth(data, BandPass, lowFrequency, highFrequency, sampleRate, order, zeroPhase)
}

func obsPyButterworth(data []float64, passband Passband, lowFrequency, highFrequency, sampleRate float64, order int, zeroPhase bool) ([]float64, error) {
	period, err := samplePeriodForRate(sampleRate)
	if err != nil {
		return nil, err
	}
	filter, err := NewButterworth(order, passband, lowFrequency, highFrequency, period)
	if err != nil {
		return nil, err
	}
	if zeroPhase {
		return filter.ApplyZeroPhase(data), nil
	}
	zeroState := filter.withZeroState()
	return zeroState.applyFloat64WithState(data, zeroState.sections), nil
}

func samplePeriodForRate(sampleRate float64) (float64, error) {
	if math.IsNaN(sampleRate) || math.IsInf(sampleRate, 0) || sampleRate <= 0 {
		return 0, fmt.Errorf("sample rate must be finite and positive")
	}
	return 1 / sampleRate, nil
}

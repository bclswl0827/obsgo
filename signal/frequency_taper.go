package signal

import (
	"fmt"
	"math"
)

func CalcFrequencyTaper(frequency, lowCut, lowPass, highPass, highCut float64) (float64, error) {
	if lowCut > lowPass || lowPass > highPass || highPass > highCut {
		return 0, fmt.Errorf("frequency corners must satisfy lowCut <= lowPass <= highPass <= highCut")
	}
	if frequency <= lowCut || frequency >= highCut {
		return 0, nil
	}
	if frequency >= lowPass && frequency <= highPass {
		return 1, nil
	}
	if frequency < lowPass {
		return 0.5 * (1 + math.Cos(math.Pi*(frequency-lowPass)/(lowCut-lowPass))), nil
	}
	return 0.5 * (1 - math.Cos(math.Pi*(frequency-highCut)/(highPass-highCut))), nil
}

// CalcFreqTaper is the SeisPlotJS-compatible function name.
func CalcFreqTaper(frequency, lowCut, lowPass, highPass, highCut float64) (float64, error) {
	return CalcFrequencyTaper(frequency, lowCut, lowPass, highPass, highCut)
}

func ApplyFrequencyTaper(result *FFTResult, lowCut, lowPass, highPass, highCut float64) (*FFTResult, error) {
	if result == nil {
		return nil, fmt.Errorf("FFT result is nil")
	}
	amplitudes, phases, err := result.AsAmpPhase()
	if err != nil {
		return nil, err
	}
	step := result.FundamentalFrequency()
	for index := range amplitudes {
		factor, err := CalcFrequencyTaper(float64(index)*step, lowCut, lowPass, highPass, highCut)
		if err != nil {
			return nil, err
		}
		amplitudes[index] *= factor
	}
	return NewFFTResultFromAmpPhase(amplitudes, phases, result.OriginalLength, result.SampleRate)
}

// ApplyFreqTaper is the SeisPlotJS-compatible function name.
func ApplyFreqTaper(result *FFTResult, sampleRate, lowCut, lowPass, highPass, highCut float64) (*FFTResult, error) {
	if result == nil {
		return nil, fmt.Errorf("FFT result is nil")
	}
	if sampleRate > 0 && math.Abs(result.SampleRate-sampleRate)/sampleRate > 0.001 {
		return nil, fmt.Errorf("FFT sample rate %g differs from %g", result.SampleRate, sampleRate)
	}
	return ApplyFrequencyTaper(result, lowCut, lowPass, highPass, highCut)
}

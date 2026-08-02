package signal

import (
	"math"
)

// ObsPyAnalyticSignal returns the FFT-based analytic signal x+iH(x), matching
// scipy.signal.hilbert and ObsPy. The transform is performed at the input
// length; no implicit zero padding is added.
func ObsPyAnalyticSignal(data []float64) []complex128 {
	if len(data) == 0 {
		return []complex128{}
	}
	spectrum := make([]complex128, len(data))
	for index, value := range data {
		spectrum[index] = complex(value, 0)
	}
	fftExactInPlace(spectrum, false)

	if len(spectrum)%2 == 0 {
		for index := 1; index < len(spectrum)/2; index++ {
			spectrum[index] *= 2
		}
		for index := len(spectrum)/2 + 1; index < len(spectrum); index++ {
			spectrum[index] = 0
		}
	} else {
		for index := 1; index <= len(spectrum)/2; index++ {
			spectrum[index] *= 2
		}
		for index := (len(spectrum) + 1) / 2; index < len(spectrum); index++ {
			spectrum[index] = 0
		}
	}
	fftExactInPlace(spectrum, true)
	return spectrum
}

// ObsPyHilbert returns the imaginary part of ObsPyAnalyticSignal.
func ObsPyHilbert(data []float64) []float64 {
	signal := ObsPyAnalyticSignal(data)
	output := make([]float64, len(signal))
	for index, value := range signal {
		output[index] = imag(value)
	}
	return output
}

// ObsPyEnvelope calculates the FFT analytic-signal envelope used by ObsPy.
func ObsPyEnvelope(data []float64) []float64 {
	analytic := ObsPyAnalyticSignal(data)
	output := make([]float64, len(data))
	for index, value := range data {
		output[index] = math.Hypot(value, imag(analytic[index]))
	}
	return output
}

// CenteredEnvelope calculates the SeisPlotJS/OregonDSP envelope. The
// centered FIR convolution has a group delay and produces a full convolution;
// only the original sample interval is returned here, matching SeisPlotJS.
func CenteredEnvelope(data []float64) []float64 {
	hilbert := CenteredHilbert(data)
	output := make([]complex128, len(data))
	for index, value := range data {
		output[index] = complex(value, hilbert[index])
	}
	result := make([]float64, len(data))
	for index, value := range output {
		result[index] = math.Hypot(real(value), imag(value))
	}
	return result
}

package signal

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/bclswl0827/obsgo/inventory"
)

type ResponseFFTMode int

const (
	// FastResponseFFT zero-pads to a power of two and uses the radix-2 FFT.
	FastResponseFFT ResponseFFTMode = iota
	// ObsPyResponseFFT follows ObsPy's _npts2nfft length selection. Non-power-
	// of-two transforms use Bluestein's algorithm and require more CPU time.
	ObsPyResponseFFT
)

type RemoveResponseOptions struct {
	// WaterLevel limits response inversion in decibels below the maximum
	// response amplitude. A negative value disables the limit.
	WaterLevel float64
	// PreFilter contains low-cut, low-pass, high-pass, and high-cut corners.
	// An all-zero value disables the frequency-domain taper.
	PreFilter [4]float64
	// TaperWidth is the total fraction tapered across both ends before the
	// transformation, matching ObsPy's taper_fraction option.
	TaperWidth float64
	// FFTMode selects the response-removal FFT strategy. The zero value uses
	// the faster power-of-two transform.
	FFTMode ResponseFFTMode
}

func RemoveResponse(data []float64, sampleRate float64, stream inventory.Stream, output ResponseOutput, options *RemoveResponseOptions) ([]float64, error) {
	results, err := RemoveResponses(data, sampleRate, stream, []ResponseOutput{output}, options)
	if err != nil {
		return nil, err
	}
	return results[output], nil
}

// RemoveResponses removes one instrument response for several requested
// motion units while sharing the time-domain taper and forward FFT.
func RemoveResponses(data []float64, sampleRate float64, stream inventory.Stream, outputs []ResponseOutput, options *RemoveResponseOptions) (map[ResponseOutput][]float64, error) {
	if sampleRate <= 0 {
		return nil, fmt.Errorf("sample rate must be positive")
	}
	if stream.Response == nil {
		return nil, fmt.Errorf("stream %q has no response", stream.Code)
	}
	response := streamResponse(stream)
	var err error
	outputUnits := make(map[ResponseOutput]string, len(outputs))
	for _, output := range outputs {
		unit, err := outputUnitName(output)
		if err != nil {
			return nil, err
		}
		outputUnits[output] = unit
	}
	configuration := RemoveResponseOptions{WaterLevel: 60, TaperWidth: 0.05}
	if options != nil {
		configuration = *options
	}
	processed := append([]float64(nil), data...)
	if configuration.TaperWidth != 0 {
		processed, err = CosineTaper(processed, configuration.TaperWidth, false, true)
		if err != nil {
			return nil, err
		}
	}
	spectrum, err := forwardResponseSpectrum(processed, configuration.FFTMode)
	if err != nil {
		return nil, err
	}
	results := make(map[ResponseOutput][]float64, len(outputs))
	for _, output := range outputs {
		results[output], err = removeResponseFromSpectrum(
			spectrum,
			len(processed),
			sampleRate,
			response,
			outputUnits[output],
			configuration,
		)
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func forwardResponseSpectrum(data []float64, mode ResponseFFTMode) ([]complex128, error) {
	if len(data) == 0 {
		return []complex128{}, nil
	}
	var size int
	switch mode {
	case FastResponseFFT:
		size, _ = FindPowerTwo(2 * len(data))
	case ObsPyResponseFFT:
		size = responseFFTLength(len(data))
	default:
		return nil, fmt.Errorf("unsupported response FFT mode %d", mode)
	}
	spectrum := make([]complex128, size)
	for index, value := range data {
		spectrum[index] = complex(value, 0)
	}
	if mode == ObsPyResponseFFT {
		fftExactInPlace(spectrum, false)
	} else {
		fftInPlace(spectrum, false)
	}
	return spectrum, nil
}

func removeResponseFromSpectrum(inputSpectrum []complex128, originalLength int, sampleRate float64, response *inventory.Response, outputUnit string, options RemoveResponseOptions) ([]float64, error) {
	if originalLength == 0 {
		return []float64{}, nil
	}
	spectrum := append([]complex128(nil), inputSpectrum...)
	size := len(spectrum)
	step := sampleRate / float64(size)
	frequencies := make([]float64, size/2+1)
	for index := range frequencies {
		frequencies[index] = float64(index) * step
	}
	responseValues, err := EvalResponse(response, frequencies, sampleRate, outputUnit)
	if err != nil {
		return nil, err
	}
	maxAmplitude := 0.0
	for _, responseValue := range responseValues {
		amplitude := cmplx.Abs(responseValue)
		if isFiniteComplex(responseValue) && amplitude > maxAmplitude {
			maxAmplitude = amplitude
		}
	}
	floor := 0.0
	if options.WaterLevel >= 0 {
		floor = maxAmplitude * math.Pow(10, -options.WaterLevel/20)
	}
	usePreFilter := options.PreFilter != [4]float64{}
	if usePreFilter {
		if _, err := CalcFrequencyTaper(0, options.PreFilter[0], options.PreFilter[1], options.PreFilter[2], options.PreFilter[3]); err != nil {
			return nil, err
		}
	}
	for index, responseValue := range responseValues {
		frequency := float64(index) * step
		amplitude := cmplx.Abs(responseValue)
		inverse := complex(0, 0)
		if amplitude > 0 && !math.IsInf(amplitude, 0) && !math.IsNaN(amplitude) {
			if floor > 0 && amplitude < floor {
				inverse = cmplx.Conj(responseValue) / complex(amplitude*floor, 0)
			} else {
				inverse = cmplx.Conj(responseValue) / complex(amplitude*amplitude, 0)
			}
		}
		if usePreFilter {
			taper, _ := CalcFrequencyTaper(frequency, options.PreFilter[0], options.PreFilter[1], options.PreFilter[2], options.PreFilter[3])
			inverse *= complex(taper, 0)
		}
		value := spectrum[index] * inverse
		if index == size/2 {
			value = complex(cmplx.Abs(value), 0)
		}
		setConjugateSpectrum(spectrum, index, value)
	}
	if options.FFTMode == ObsPyResponseFFT {
		fftExactInPlace(spectrum, true)
	} else {
		fftInPlace(spectrum, true)
	}
	output := make([]float64, originalLength)
	for index := range output {
		output[index] = real(spectrum[index])
	}
	return output, nil
}

func complexPower(value complex128, power int) complex128 {
	result := complex(1, 0)
	for index := 0; index < power; index++ {
		result *= value
	}
	return result
}

func setConjugateSpectrum(spectrum []complex128, index int, value complex128) {
	if index == 0 || index == len(spectrum)/2 {
		spectrum[index] = complex(real(value), 0)
		return
	}
	spectrum[index] = value
	spectrum[len(spectrum)-index] = cmplx.Conj(value)
}

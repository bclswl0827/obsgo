package signal

import (
	"fmt"

	"github.com/bclswl0827/obsgo/inventory"
)

// SimulateResponseOptions controls frequency-domain instrument simulation.
// The simulated output is in the final response-stage output units, normally
// counts. TaperWidth is applied before the FFT; PreFilter is applied to the
// simulated spectrum.
type SimulateResponseOptions struct {
	PreFilter  [4]float64
	TaperWidth float64
	FFTMode    ResponseFFTMode
}

// SimulateResponse applies an instrument response to physical input data.
// Input selects the displacement, velocity, or acceleration unit of data.
// The output is expressed in the response's final stage output units.
func SimulateResponse(data []float64, sampleRate float64, response *inventory.Response, input ResponseOutput, options *SimulateResponseOptions) ([]float64, error) {
	if sampleRate <= 0 {
		return nil, fmt.Errorf("sample rate must be positive")
	}
	inputUnit, err := outputUnitName(input)
	if err != nil {
		return nil, err
	}
	configuration := SimulateResponseOptions{}
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
	if len(processed) == 0 {
		return []float64{}, nil
	}
	size := len(spectrum)
	frequencies := make([]float64, size/2+1)
	step := sampleRate / float64(size)
	for index := range frequencies {
		frequencies[index] = float64(index) * step
	}
	responseValues, err := EvalResponse(response, frequencies, sampleRate, inputUnit)
	if err != nil {
		return nil, err
	}
	usePreFilter := configuration.PreFilter != [4]float64{}
	if usePreFilter {
		if _, err := CalcFrequencyTaper(0, configuration.PreFilter[0], configuration.PreFilter[1], configuration.PreFilter[2], configuration.PreFilter[3]); err != nil {
			return nil, err
		}
	}
	for index, responseValue := range responseValues {
		if usePreFilter {
			taper, _ := CalcFrequencyTaper(frequencies[index], configuration.PreFilter[0], configuration.PreFilter[1], configuration.PreFilter[2], configuration.PreFilter[3])
			responseValue *= complex(taper, 0)
		}
		value := spectrum[index] * responseValue
		if !isFiniteComplex(value) {
			value = 0
		}
		setConjugateSpectrum(spectrum, index, value)
	}
	if configuration.FFTMode == ObsPyResponseFFT {
		fftExactInPlace(spectrum, true)
	} else {
		fftInPlace(spectrum, true)
	}
	output := make([]float64, len(processed))
	for index := range output {
		output[index] = real(spectrum[index])
	}
	return output, nil
}

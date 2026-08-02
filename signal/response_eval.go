package signal

import (
	"fmt"
	"math"
	"sort"

	"github.com/bclswl0827/obsgo/inventory"
)

// EvalResponseOptions selects a contiguous range of response stages. A zero
// value evaluates the complete response. Stage numbers are StationXML stage
// numbers, not indexes into Response.Stages.
type EvalResponseOptions struct {
	StartStage int
	EndStage   int
}

// EvalResponse evaluates the complete instrument response at non-negative
// frequencies. The returned transfer function is expressed as final-stage
// output units per one unit of the requested physical input (for example,
// counts per m/s2).
func EvalResponse(response *inventory.Response, frequencies []float64, sampleRate float64, output string) ([]complex128, error) {
	return EvalResponseWithOptions(response, frequencies, sampleRate, output, EvalResponseOptions{})
}

// EvalResponseAt evaluates one frequency of a complete instrument response.
func EvalResponseAt(response *inventory.Response, frequency, sampleRate float64, output string) (complex128, error) {
	values, err := EvalResponse(response, []float64{frequency}, sampleRate, output)
	if err != nil {
		return 0, err
	}
	return values[0], nil
}

// EvalResponseWithOptions is EvalResponse with optional stage selection.
func EvalResponseWithOptions(response *inventory.Response, frequencies []float64, sampleRate float64, output string, options EvalResponseOptions) ([]complex128, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}
	if sampleRate <= 0 {
		return nil, fmt.Errorf("sample rate must be positive")
	}
	for _, frequency := range frequencies {
		if frequency < 0 || math.IsNaN(frequency) || math.IsInf(frequency, 0) {
			return nil, fmt.Errorf("frequency must be finite and non-negative")
		}
	}
	stages := responseStages(response, options)
	if len(response.Stages) != 0 && len(stages) == 0 {
		return nil, fmt.Errorf("response stage range does not match any stage")
	}
	inputUnit := parseMotionUnit(responseInputUnit(stages, response))
	targetUnit := parseMotionUnit(output)
	result := make([]complex128, len(frequencies))
	for index, frequency := range frequencies {
		value, err := evalResponseChain(response, stages, frequency, sampleRate)
		if err != nil {
			return nil, err
		}
		result[index] = convertResponseUnit(value, frequency, inputUnit, targetUnit)
	}
	return result, nil
}

// CalcResponseAtFrequencies is a descriptive alias for EvalResponse.
func CalcResponseAtFrequencies(response *inventory.Response, frequencies []float64, sampleRate float64, output string) ([]complex128, error) {
	return EvalResponse(response, frequencies, sampleRate, output)
}

type responseStage struct {
	stage inventory.Stage
}

func responseStages(response *inventory.Response, options EvalResponseOptions) []responseStage {
	if len(response.Stages) == 0 {
		return nil
	}
	stages := append([]inventory.Stage(nil), response.Stages...)
	sort.SliceStable(stages, func(left, right int) bool {
		return stages[left].Number < stages[right].Number
	})
	selected := make([]responseStage, 0, len(stages))
	for _, stage := range stages {
		if options.StartStage != 0 && stage.Number < options.StartStage {
			continue
		}
		if options.EndStage != 0 && stage.Number > options.EndStage {
			continue
		}
		selected = append(selected, responseStage{stage: stage})
	}
	return selected
}

func evalResponseChain(response *inventory.Response, stages []responseStage, frequency, sampleRate float64) (complex128, error) {
	if len(stages) == 0 {
		return evalLegacyResponse(response, frequency, sampleRate)
	}
	value := complex(1, 0)
	hasGain := false
	rates := responseStageInputRates(stages, sampleRate)
	for stageIndex, item := range stages {
		stage := item.stage
		stageRate := rates[stageIndex]
		if stageRate <= 0 {
			return 0, fmt.Errorf("response stage %d has no positive input sample rate", stage.Number)
		}
		if stage.PolesZeros != nil {
			part, err := evalPolesZeros(stage.PolesZeros, frequency, stageRate)
			if err != nil {
				return 0, fmt.Errorf("evaluate response stage %d pole-zero: %w", stage.Number, err)
			}
			value *= part
		}
		if stage.Coefficients != nil {
			part, err := evalCoefficients(stage.Coefficients, frequency, stageRate)
			if err != nil {
				return 0, fmt.Errorf("evaluate response stage %d coefficients: %w", stage.Number, err)
			}
			value *= part
		}
		if stage.FIR != nil {
			part, err := evalFIR(stage.FIR, frequency, stageRate)
			if err != nil {
				return 0, fmt.Errorf("evaluate response stage %d FIR: %w", stage.Number, err)
			}
			value *= part
		}
		if stage.Gain != nil {
			value *= complex(stage.Gain.Value, 0)
			hasGain = true
		}
		if stage.Decimation != nil {
			asymmetricFIR := stage.FIR != nil && !isSymmetricFIR(stage.FIR)
			part, _, err := evalDecimation(stage.Decimation, frequency, asymmetricFIR)
			if err != nil {
				return 0, fmt.Errorf("evaluate response stage %d decimation: %w", stage.Number, err)
			}
			value *= part
		}
	}
	// InstrumentSensitivity is a summary of the complete stage chain. It is
	// used only when the chain did not provide stage gains, preventing the
	// common sensitivity-times-stage-gains double multiplication.
	if !hasGain {
		gain := response.Gain
		if response.Sensitivity != nil && response.Sensitivity.Value != 0 {
			gain = response.Sensitivity.Value
		}
		if gain != 0 {
			value *= complex(gain, 0)
		}
	}
	return value, nil
}

func responseStageInputRates(stages []responseStage, sampleRate float64) []float64 {
	rates := make([]float64, len(stages))
	currentRate := sampleRate
	for _, item := range stages {
		if item.stage.Decimation != nil && item.stage.Decimation.InputSampleRate > 0 {
			currentRate = item.stage.Decimation.InputSampleRate
			break
		}
	}
	for index, item := range stages {
		stage := item.stage
		if stage.Decimation != nil && stage.Decimation.InputSampleRate > 0 {
			currentRate = stage.Decimation.InputSampleRate
		}
		rates[index] = currentRate
		factor := 1
		if stage.Decimation != nil && stage.Decimation.Factor > 0 {
			factor = stage.Decimation.Factor
		}
		if stage.Decimation != nil && stage.Decimation.InputSampleRate > 0 {
			currentRate = stage.Decimation.InputSampleRate / float64(factor)
		} else {
			currentRate /= float64(factor)
		}
	}
	return rates
}

func evalLegacyResponse(response *inventory.Response, frequency, sampleRate float64) (complex128, error) {
	if len(response.Poles) == 0 && len(response.Zeros) == 0 && response.Type == "" {
		gain := response.Gain
		if response.Sensitivity != nil && response.Sensitivity.Value != 0 {
			gain = response.Sensitivity.Value
		}
		return complex(gain, 0), nil
	}
	value := &inventory.PolesZeros{
		TransferFunctionType:   response.Type,
		NormalizationFactor:    response.NormalizationFactor,
		NormalizationFrequency: response.NormalizationFrequency,
		Zeros:                  response.Zeros,
		Poles:                  response.Poles,
	}
	result, err := evalPolesZeros(value, frequency, sampleRate)
	if err != nil {
		return 0, err
	}
	gain := response.Gain
	if response.Sensitivity != nil && response.Sensitivity.Value != 0 {
		gain = response.Sensitivity.Value
	}
	return result * complex(gain, 0), nil
}

func convertResponseUnit(value complex128, frequency float64, input, target motionUnit) complex128 {
	if !input.known || !target.known {
		return value
	}
	value *= complex(target.scale/input.scale, 0)
	power := input.gamma - target.gamma
	if power >= 0 {
		return value * complexPower(complex(0, 2*math.Pi*frequency), power)
	}
	divisor := complexPower(complex(0, 2*math.Pi*frequency), -power)
	if divisor == 0 {
		return 0
	}
	return value / divisor
}

func isFiniteComplex(value complex128) bool {
	return !math.IsNaN(real(value)) && !math.IsNaN(imag(value)) && !math.IsInf(real(value), 0) && !math.IsInf(imag(value), 0)
}

package signal

import (
	"fmt"
	"math"
)

type biquad struct {
	b0 float64
	b1 float64
	b2 float64
	a1 float64
	a2 float64
	s1 float64
	s2 float64
}

// IIRFilter is a stateful cascaded second-order-section filter.
type IIRFilter struct {
	Numerator    []float64
	Denominator  []float64
	SamplePeriod float64
	sections     []biquad
}

// ZeroPhasePadType selects the endpoint extension used by ApplyZeroPhaseFiltfilt.
// ApplyZeroPhase itself intentionally does not pad, matching ObsPy's public
// zerophase filter helpers.
type ZeroPhasePadType int

const (
	OddPadding ZeroPhasePadType = iota
	EvenPadding
	ConstantPadding
)

// ZeroPhaseFiltfiltOptions controls the SciPy-style padded zero-phase pass.
// A non-positive PadLength uses the default SOS pad length.
type ZeroPhaseFiltfiltOptions struct {
	PadType   ZeroPhasePadType
	PadLength int
}

// Apply filters a copy of input and retains state for the next call.
func (filter *IIRFilter) Apply(input []float64) []float64 {
	output := append([]float64(nil), input...)
	filter.ApplyInPlace(output)
	return output
}

func (filter *IIRFilter) ApplyInPlace(data []float64) {
	for sectionIndex := range filter.sections {
		section := &filter.sections[sectionIndex]
		for index, input := range data {
			// OregonDSP processes FloatArray values. Keep the public Go API as
			// []float64, but preserve the Float32 input/output rounding and its
			// direct-form-II state equation here.
			x := float64(float32(input))
			s0 := x - section.a1*section.s1 - section.a2*section.s2
			output := section.b0*s0 + section.b1*section.s1 + section.b2*section.s2
			section.s2 = section.s1
			section.s1 = s0
			data[index] = float64(float32(output))
		}
	}
}

// ApplyZeroPhase applies the filter forwards and backwards with independent
// zero state. This is the behavior of ObsPy's signal.filter zerophase=True:
// sosfilt(data), reverse, sosfilt(data), reverse. It does not use endpoint
// padding and therefore also works for short records.
func (filter *IIRFilter) ApplyZeroPhase(input []float64) []float64 {
	forwardFilter := filter.withZeroState()
	forward := forwardFilter.applyFloat64WithState(input, forwardFilter.sections)
	reverse(forward)
	backwardFilter := filter.withZeroState()
	output := backwardFilter.applyFloat64WithState(forward, backwardFilter.sections)
	reverse(output)
	return output
}

// ApplyZeroPhaseFiltfilt applies the filter using SciPy's sosfiltfilt-style
// endpoint extension and steady-state initial conditions. It is separate from
// ApplyZeroPhase because ObsPy's public zero-phase helpers use an unpadded
// forward/reverse pass.
func (filter *IIRFilter) ApplyZeroPhaseFiltfilt(input []float64, options ZeroPhaseFiltfiltOptions) ([]float64, error) {
	if filter == nil {
		return nil, fmt.Errorf("IIR filter is nil")
	}
	if len(input) == 0 {
		return []float64{}, nil
	}
	if len(filter.sections) == 0 {
		return nil, fmt.Errorf("IIR filter has no sections")
	}
	edge := options.PadLength
	if edge <= 0 {
		edge = filter.defaultFiltfiltPadLength()
	}
	if edge < 0 {
		return nil, fmt.Errorf("filtfilt pad length must not be negative")
	}
	if len(input) <= edge {
		return nil, fmt.Errorf("filtfilt input length %d must be greater than pad length %d", len(input), edge)
	}

	extended, err := padForFiltfilt(input, edge, options.PadType)
	if err != nil {
		return nil, err
	}
	initial := filter.steadyState(edgeValue(extended))
	forward := filter.applyFloat64WithState(extended, initial)
	reverse(forward)
	initial = filter.steadyState(edgeValue(forward))
	output := filter.applyFloat64WithState(forward, initial)
	reverse(output)
	return append([]float64(nil), output[edge:len(output)-edge]...), nil
}

func (filter *IIRFilter) defaultFiltfiltPadLength() int {
	zeroNumerator := 0
	zeroDenominator := 0
	for _, section := range filter.sections {
		if section.b2 == 0 {
			zeroNumerator++
		}
		if section.a2 == 0 {
			zeroDenominator++
		}
	}
	ntaps := 2*len(filter.sections) + 1
	if zeroNumerator < zeroDenominator {
		ntaps -= zeroNumerator
	} else {
		ntaps -= zeroDenominator
	}
	return 3 * ntaps
}

func (filter *IIRFilter) steadyState(firstSample float64) []biquad {
	state := make([]biquad, len(filter.sections))
	level := 1.0
	for index, section := range filter.sections {
		denominator := 1 + section.a1 + section.a2
		if denominator == 0 {
			state[index] = section
			continue
		}
		sectionState := level / denominator
		section.s1 = sectionState * firstSample
		section.s2 = sectionState * firstSample
		state[index] = section
		level = (section.b0 + section.b1 + section.b2) * sectionState
	}
	return state
}

func (filter *IIRFilter) applyFloat64WithState(data []float64, state []biquad) []float64 {
	output := append([]float64(nil), data...)
	for sectionIndex := range filter.sections {
		section := state[sectionIndex]
		for index, input := range output {
			s0 := input - section.a1*section.s1 - section.a2*section.s2
			value := section.b0*s0 + section.b1*section.s1 + section.b2*section.s2
			section.s2 = section.s1
			section.s1 = s0
			output[index] = value
		}
	}
	return output
}

func edgeValue(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	return data[0]
}

func padForFiltfilt(data []float64, edge int, padType ZeroPhasePadType) ([]float64, error) {
	if edge == 0 {
		return append([]float64(nil), data...), nil
	}
	if edge >= len(data) {
		return nil, fmt.Errorf("filtfilt pad length %d must be smaller than input length %d", edge, len(data))
	}
	left := make([]float64, edge)
	right := make([]float64, edge)
	for index := 0; index < edge; index++ {
		switch padType {
		case OddPadding:
			left[index] = 2*data[0] - data[edge-index]
			right[index] = 2*data[len(data)-1] - data[len(data)-2-index]
		case EvenPadding:
			left[index] = data[edge-index]
			right[index] = data[len(data)-2-index]
		case ConstantPadding:
			left[index] = data[0]
			right[index] = data[len(data)-1]
		default:
			return nil, fmt.Errorf("unsupported filtfilt pad type %d", padType)
		}
	}
	output := make([]float64, 0, len(data)+2*edge)
	output = append(output, left...)
	output = append(output, data...)
	output = append(output, right...)
	return output, nil
}

func (filter *IIRFilter) Reset() {
	for index := range filter.sections {
		filter.sections[index].s1 = 0
		filter.sections[index].s2 = 0
	}
}

func (filter *IIRFilter) withZeroState() *IIRFilter {
	clone := *filter
	clone.Numerator = append([]float64(nil), filter.Numerator...)
	clone.Denominator = append([]float64(nil), filter.Denominator...)
	clone.sections = append([]biquad(nil), filter.sections...)
	clone.Reset()
	return &clone
}

func reverse(values []float64) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func GetPassband(value string) (Passband, error) {
	switch value {
	case "LOWPASS":
		return LowPass, nil
	case "BANDPASS":
		return BandPass, nil
	case "HIGHPASS":
		return HighPass, nil
	default:
		return LowPass, fmt.Errorf("unknown passband %q", value)
	}
}

func ApplyFilter(filter *IIRFilter, data []float64, samplePeriod float64) ([]float64, error) {
	if filter == nil {
		return nil, fmt.Errorf("IIR filter is nil")
	}
	if math.IsNaN(samplePeriod) || math.IsInf(samplePeriod, 0) || samplePeriod <= 0 {
		return nil, fmt.Errorf("sample period must be finite and positive")
	}
	if math.IsNaN(filter.SamplePeriod) || math.IsInf(filter.SamplePeriod, 0) || filter.SamplePeriod <= 0 {
		return nil, fmt.Errorf("filter sample period must be finite and positive")
	}
	if math.Abs(filter.SamplePeriod-samplePeriod)/samplePeriod > 0.001 {
		return nil, fmt.Errorf("filter sample period %g differs from data sample period %g", filter.SamplePeriod, samplePeriod)
	}
	return filter.Apply(data), nil
}

package signal

import (
	"fmt"
	"math"
)

const correlationEpsilon = 2.220446049250313e-16

// CorrelationMethod selects the implementation used by ObsPy-style
// correlation. Auto chooses direct correlation for small inputs and FFT for
// larger inputs.
type CorrelationMethod string

const (
	CorrelationAuto   CorrelationMethod = "auto"
	CorrelationDirect CorrelationMethod = "direct"
	CorrelationFFT    CorrelationMethod = "fft"
)

// CorrelationNormalization selects the normalization used by correlation.
// Full normalization is available for template correlation only.
type CorrelationNormalization string

const (
	CorrelationNaive CorrelationNormalization = "naive"
	CorrelationFull  CorrelationNormalization = "full"
	CorrelationNone  CorrelationNormalization = "none"
)

// CorrelationOptions is the explicit equivalent of ObsPy correlate's
// demean, normalize, and method arguments.
type CorrelationOptions struct {
	Demean    bool
	Normalize CorrelationNormalization
	Method    CorrelationMethod
}

// DefaultCorrelationOptions returns ObsPy correlate's defaults.
func DefaultCorrelationOptions() CorrelationOptions {
	return CorrelationOptions{
		Demean:    true,
		Normalize: CorrelationNaive,
		Method:    CorrelationAuto,
	}
}

// ObsPyCorrelate uses ObsPy correlate's default demean, naive normalization,
// and automatic implementation selection.
func ObsPyCorrelate(a, b []float64, shift int) ([]float64, error) {
	return Correlate(a, b, shift, DefaultCorrelationOptions())
}

// Correlate calculates the ObsPy correlate result centered around zero lag.
// A non-negative shift returns 2*shift+1 samples for an odd full correlation
// length and 2*shift samples for an even full correlation length. A negative
// shift returns the complete correlation. Different length inputs are aligned
// around their middle, as in ObsPy.
func Correlate(a, b []float64, shift int, options CorrelationOptions) ([]float64, error) {
	if len(a) == 0 || len(b) == 0 {
		return []float64{}, nil
	}
	if shift < 0 {
		shift = (len(a) + len(b) - 1) / 2
	}
	if options.Normalize == "" {
		options.Normalize = CorrelationNaive
	}
	if options.Method == "" {
		options.Method = CorrelationAuto
	}
	if options.Normalize != CorrelationNaive && options.Normalize != CorrelationNone {
		return nil, fmt.Errorf("correlate normalization must be naive or none")
	}
	a, b = correlationPreprocess(a, b, options.Demean)
	full, err := correlationFull(a, b, options.Method)
	if err != nil {
		return nil, err
	}
	if options.Normalize == CorrelationNaive {
		norm := math.Sqrt(sumSquares(a) * sumSquares(b))
		if norm <= correlationEpsilon {
			for index := range full {
				full[index] = 0
			}
		} else {
			for index := range full {
				full[index] /= norm
			}
		}
	}
	return centeredCorrelation(full, shift), nil
}

// CorrelationLags returns the fractional sample lags corresponding to
// Correlate's centered output. Even-length outputs have a half-sample center,
// matching ObsPy xcorr_max.
func CorrelationLags(length int) []float64 {
	if length <= 0 {
		return []float64{}
	}
	center := float64(length-1) / 2
	lags := make([]float64, length)
	for index := range lags {
		lags[index] = float64(index) - center
	}
	return lags
}

// CorrelationPeakObsPy returns the lag and value of the largest absolute
// value, using ObsPy's fractional center for even-length outputs.
func CorrelationPeakObsPy(values []float64, absolute bool) (lag, value float64, err error) {
	if len(values) == 0 {
		return 0, 0, fmt.Errorf("correlation values must not be empty")
	}
	best := 0
	for index := 1; index < len(values); index++ {
		better := values[index] > values[best]
		if absolute {
			better = math.Abs(values[index]) > math.Abs(values[best])
		}
		if better {
			best = index
		}
	}
	return CorrelationLags(len(values))[best], values[best], nil
}

// TemplateCorrelationMode selects the output shape of CorrelateTemplate.
type TemplateCorrelationMode string

const (
	TemplateValid TemplateCorrelationMode = "valid"
	TemplateSame  TemplateCorrelationMode = "same"
	TemplateFull  TemplateCorrelationMode = "full"
)

// TemplateCorrelationOptions is the explicit equivalent of ObsPy
// correlate_template's mode, normalize, demean, and method arguments.
type TemplateCorrelationOptions struct {
	Mode      TemplateCorrelationMode
	Normalize CorrelationNormalization
	Demean    bool
	Method    CorrelationMethod
}

// DefaultTemplateCorrelationOptions returns ObsPy's default template
// correlation configuration: valid mode, full normalization, and demean.
func DefaultTemplateCorrelationOptions() TemplateCorrelationOptions {
	return TemplateCorrelationOptions{
		Mode:      TemplateValid,
		Normalize: CorrelationFull,
		Demean:    true,
		Method:    CorrelationAuto,
	}
}

// ObsPyCorrelateTemplate uses ObsPy correlate_template's default options.
func ObsPyCorrelateTemplate(data, template []float64) ([]float64, error) {
	return CorrelateTemplate(data, template, DefaultTemplateCorrelationOptions())
}

// CorrelateTemplate calculates a moving template correlation. Full
// normalization computes the mean and energy of each data window, including
// zero padding for full and same modes, exactly as ObsPy does.
func CorrelateTemplate(data, template []float64, options TemplateCorrelationOptions) ([]float64, error) {
	if len(template) == 0 {
		return nil, fmt.Errorf("template must not be empty")
	}
	if len(data) < len(template) {
		return nil, fmt.Errorf("data must not be shorter than template")
	}
	if options.Mode == "" {
		options.Mode = TemplateValid
	}
	if options.Normalize == "" {
		options.Normalize = CorrelationFull
	}
	if options.Method == "" {
		options.Method = CorrelationAuto
	}
	if options.Mode != TemplateValid && options.Mode != TemplateSame && options.Mode != TemplateFull {
		return nil, fmt.Errorf("unsupported template correlation mode %q", options.Mode)
	}
	if options.Normalize != CorrelationNone && options.Normalize != CorrelationNaive && options.Normalize != CorrelationFull {
		return nil, fmt.Errorf("unsupported template correlation normalization %q", options.Normalize)
	}
	centeredTemplate := append([]float64(nil), template...)
	if options.Demean {
		centeredTemplate = RemoveMean(centeredTemplate)
	}
	correlationData := append([]float64(nil), data...)
	if options.Demean && options.Normalize != CorrelationFull {
		correlationData = RemoveMean(correlationData)
	}
	full, err := correlationFull(correlationData, centeredTemplate, options.Method)
	if err != nil {
		return nil, err
	}
	start, count := templateCorrelationSlice(len(data), len(template), options.Mode)
	result := append([]float64(nil), full[start:start+count]...)
	if options.Normalize == CorrelationNone {
		return result, nil
	}
	templateEnergy := sumSquares(centeredTemplate)
	if options.Normalize == CorrelationNaive {
		norm := math.Sqrt(templateEnergy * sumSquares(correlationData))
		if norm <= correlationEpsilon {
			for index := range result {
				result[index] = 0
			}
		} else {
			for index := range result {
				result[index] /= norm
			}
		}
		return result, nil
	}
	for outputIndex := range result {
		fullIndex := start + outputIndex
		dataStart := fullIndex - (len(template) - 1)
		var sum, sumSquaresWindow float64
		for index := 0; index < len(template); index++ {
			dataIndex := dataStart + index
			if dataIndex < 0 || dataIndex >= len(correlationData) {
				continue
			}
			value := correlationData[dataIndex]
			sum += value
			sumSquaresWindow += value * value
		}
		if options.Demean {
			sumSquaresWindow -= sum * sum / float64(len(template))
		}
		if sumSquaresWindow < 0 && sumSquaresWindow > -1e-12 {
			sumSquaresWindow = 0
		}
		norm := math.Sqrt(math.Max(0, sumSquaresWindow*templateEnergy))
		if norm <= correlationEpsilon {
			result[outputIndex] = 0
		} else {
			result[outputIndex] /= norm
		}
	}
	return result, nil
}

func correlationPreprocess(a, b []float64, demean bool) ([]float64, []float64) {
	if !demean {
		return append([]float64(nil), a...), append([]float64(nil), b...)
	}
	return RemoveMean(a), RemoveMean(b)
}

func correlationFull(a, b []float64, method CorrelationMethod) ([]float64, error) {
	if method == CorrelationAuto {
		if len(a)*len(b) <= 4096 {
			method = CorrelationDirect
		} else {
			method = CorrelationFFT
		}
	}
	switch method {
	case CorrelationDirect:
		return correlationFullDirect(a, b), nil
	case CorrelationFFT:
		return correlationFullFFT(a, b), nil
	default:
		return nil, fmt.Errorf("unsupported correlation method %q", method)
	}
}

func correlationFullDirect(a, b []float64) []float64 {
	result := make([]float64, len(a)+len(b)-1)
	for left, value := range a {
		for right := range b {
			result[left+len(b)-1-right] += value * b[right]
		}
	}
	return result
}

func correlationFullFFT(a, b []float64) []float64 {
	length := len(a) + len(b) - 1
	fftLength, _ := FindPowerTwo(length)
	left := make([]complex128, fftLength)
	right := make([]complex128, fftLength)
	for index, value := range a {
		left[index] = complex(value, 0)
	}
	for index, value := range b {
		right[len(b)-1-index] = complex(value, 0)
	}
	fftInPlace(left, false)
	fftInPlace(right, false)
	for index := range left {
		left[index] *= right[index]
	}
	fftInPlace(left, true)
	result := make([]float64, length)
	for index := range result {
		result[index] = real(left[index])
	}
	return result
}

func centeredCorrelation(full []float64, shift int) []float64 {
	center := len(full) / 2
	count := 2*shift + len(full)%2
	result := make([]float64, count)
	for index := range result {
		fullIndex := center - shift + index
		if fullIndex >= 0 && fullIndex < len(full) {
			result[index] = full[fullIndex]
		}
	}
	return result
}

func templateCorrelationSlice(dataLength, templateLength int, mode TemplateCorrelationMode) (start, count int) {
	fullLength := dataLength + templateLength - 1
	switch mode {
	case TemplateValid:
		return templateLength - 1, dataLength - templateLength + 1
	case TemplateSame:
		count = dataLength
		return (fullLength - count) / 2, count
	default:
		return 0, fullLength
	}
}

func sumSquares(values []float64) float64 {
	var result float64
	for _, value := range values {
		result += value * value
	}
	return result
}

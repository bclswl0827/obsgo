package signal

import "fmt"

// Differentiate uses the first forward difference and returns n-1 samples.
func Differentiate(data []float64, sampleRate float64) ([]float64, error) {
	if sampleRate <= 0 {
		return nil, fmt.Errorf("sample rate must be positive")
	}
	if len(data) < 2 {
		return []float64{}, nil
	}
	output := make([]float64, len(data)-1)
	for index := range output {
		output[index] = (data[index+1] - data[index]) * sampleRate
	}
	return output, nil
}

// DifferentiateGradient matches ObsPy's default gradient differentiation:
// first differences at the boundaries and centered differences in the
// interior. It preserves the input length.
func DifferentiateGradient(data []float64, sampleRate float64) ([]float64, error) {
	if sampleRate <= 0 {
		return nil, fmt.Errorf("sample rate must be positive")
	}
	if len(data) == 0 {
		return []float64{}, nil
	}
	if len(data) == 1 {
		return []float64{0}, nil
	}
	output := make([]float64, len(data))
	output[0] = (data[1] - data[0]) * sampleRate
	for index := 1; index < len(data)-1; index++ {
		output[index] = 0.5 * (data[index+1] - data[index-1]) * sampleRate
	}
	output[len(output)-1] = (data[len(data)-1] - data[len(data)-2]) * sampleRate
	return output, nil
}

// Integrate reproduces the SeisPlotJS rectangular integration and returns n+1 samples.
func Integrate(data []float64, sampleRate, integrationConstant float64) ([]float64, error) {
	if sampleRate <= 0 {
		return nil, fmt.Errorf("sample rate must be positive")
	}
	output := make([]float64, len(data)+1)
	output[0] = integrationConstant
	period := 1 / sampleRate
	for index, value := range data {
		output[index+1] = output[index] + period*value
	}
	return output, nil
}

// CumulativeTrapezoid matches ObsPy's cumulative_trapezoid(initial=0).
func CumulativeTrapezoid(data []float64, sampleRate float64) ([]float64, error) {
	if sampleRate <= 0 {
		return nil, fmt.Errorf("sample rate must be positive")
	}
	output := make([]float64, len(data))
	period := 1 / sampleRate
	for index := 1; index < len(data); index++ {
		output[index] = output[index-1] + 0.5*(data[index-1]+data[index])*period
	}
	return output, nil
}

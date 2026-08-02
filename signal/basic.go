package signal

import (
	"fmt"
	"math"
)

func Amplitude(realPart, imaginaryPart float64) float64 {
	return math.Hypot(realPart, imaginaryPart)
}

func Mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	var sum float64
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}

// RemoveMean returns a copy with its arithmetic mean removed.
func RemoveMean(data []float64) []float64 {
	mean := Mean(data)
	output := make([]float64, len(data))
	for index, value := range data {
		output[index] = value - mean
	}
	return output
}

// RMean is the SeisPlotJS-compatible name for RemoveMean.
func RMean(data []float64) []float64 { return RemoveMean(data) }

func LineFit(data []float64, sampleRate float64) (LineFitResult, error) {
	if len(data) == 0 {
		return LineFitResult{}, fmt.Errorf("line fit requires at least one sample")
	}
	if sampleRate <= 0 {
		return LineFitResult{}, fmt.Errorf("sample rate must be positive")
	}
	rn := float64(len(data))
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for index, value := range data {
		x := float64(index) / sampleRate
		sumX += x
		sumY += value
		sumXY += x * value
		sumX2 += x * x
		sumY2 += value * value
	}
	d := rn*sumX2 - sumX*sumX
	result := LineFitResult{}
	if d != 0 {
		result.Intercept = (sumX2*sumY - sumX*sumXY) / d
		result.Slope = (rn*sumXY - sumX*sumY) / d
	}
	if len(data) > 1 {
		sigma2 := (sumY2 + rn*result.Intercept*result.Intercept + result.Slope*result.Slope*sumX2 -
			2*result.Intercept*sumY - 2*result.Slope*sumXY + 2*result.Intercept*result.Slope*sumX) / (rn - 1)
		if sigma2 < 0 && sigma2 > -1e-15 {
			sigma2 = 0
		}
		result.Sigma = math.Sqrt(sigma2)
		if d != 0 {
			result.SigmaSlope = math.Sqrt(rn * sigma2 / d)
			result.SigmaOffset = math.Sqrt(sigma2 * sumX2 / d)
			correlationDenominator := math.Sqrt(d * (rn*sumY2 - sumY*sumY))
			if correlationDenominator != 0 {
				result.Correlation = math.Abs((rn*sumXY - sumX*sumY) / correlationDenominator)
			}
		}
	}
	return result, nil
}

// RemoveTrend subtracts the least-squares linear trend from a copy of data.
func RemoveTrend(data []float64, sampleRate float64) ([]float64, error) {
	fit, err := LineFit(data, sampleRate)
	if err != nil {
		return nil, err
	}
	output := make([]float64, len(data))
	for index, value := range data {
		output[index] = value - fit.Intercept - fit.Slope*float64(index)/sampleRate
	}
	return output, nil
}

func Multiply(data []float64, factor float64) []float64 {
	output := make([]float64, len(data))
	for index, value := range data {
		output[index] = value * factor
	}
	return output
}

// Mul is the SeisPlotJS-compatible name for Multiply.
func Mul(data []float64, factor float64) []float64 { return Multiply(data, factor) }

func Add(data []float64, value float64) []float64 {
	output := make([]float64, len(data))
	for index, sample := range data {
		output[index] = sample + value
	}
	return output
}

func GainCorrect(data []float64, sensitivity float64) ([]float64, error) {
	if sensitivity == 0 {
		return nil, fmt.Errorf("instrument sensitivity must not be zero")
	}
	return Multiply(data, 1/sensitivity), nil
}

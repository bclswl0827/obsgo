package signal

import (
	"fmt"
	"math"
)

func NewChebyshevI(order int, epsilon float64, passband Passband, lowFrequency, highFrequency, samplePeriod float64) (*IIRFilter, error) {
	if order <= 0 {
		return nil, fmt.Errorf("filter order must be positive")
	}
	if math.IsNaN(epsilon) || math.IsInf(epsilon, 0) || epsilon <= 0 {
		return nil, fmt.Errorf("epsilon must be finite and positive")
	}
	return newOregonIIRFilter(
		newOregonChebyshevI(order, epsilon),
		passband,
		lowFrequency,
		highFrequency,
		samplePeriod,
	)
}

func CreateChebyshevI(order int, epsilon float64, passband Passband, lowFrequency, highFrequency, samplePeriod float64) (*IIRFilter, error) {
	return NewChebyshevI(order, epsilon, passband, lowFrequency, highFrequency, samplePeriod)
}

func NewChebyshevII(order int, epsilon float64, passband Passband, lowFrequency, highFrequency, samplePeriod float64) (*IIRFilter, error) {
	if order <= 0 {
		return nil, fmt.Errorf("filter order must be positive")
	}
	if math.IsNaN(epsilon) || math.IsInf(epsilon, 0) || epsilon <= 0 {
		return nil, fmt.Errorf("epsilon must be finite and positive")
	}
	return newOregonIIRFilter(
		newOregonChebyshevII(order, epsilon),
		passband,
		lowFrequency,
		highFrequency,
		samplePeriod,
	)
}

func CreateChebyshevII(order int, epsilon float64, passband Passband, lowFrequency, highFrequency, samplePeriod float64) (*IIRFilter, error) {
	return NewChebyshevII(order, epsilon, passband, lowFrequency, highFrequency, samplePeriod)
}

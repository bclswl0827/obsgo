package signal

import (
	"fmt"
	"math"
)

// CrossCorrelation is the integer-lag wrapper around ObsPy-style Correlate.
// It demeans both inputs and uses naive normalization when normalize is true.
// Use Correlate directly when fractional half-sample lags are needed.
func CrossCorrelation(a, b []float64, maxLag int, normalize bool) (lags []int, values []float64, err error) {
	options := DefaultCorrelationOptions()
	if !normalize {
		options.Normalize = CorrelationNone
	}
	values, err = Correlate(a, b, maxLag, options)
	if err != nil {
		return nil, nil, err
	}
	lags = make([]int, len(values))
	for index, lag := range CorrelationLags(len(values)) {
		lags[index] = int(math.Round(lag))
	}
	return lags, values, nil
}

// NormalizedCrossCorrelation is the normalized form of CrossCorrelation.
func NormalizedCrossCorrelation(a, b []float64, maxLag int) (lags []int, values []float64, err error) {
	return CrossCorrelation(a, b, maxLag, true)
}

// CorrelationPeak returns the lag and value of the largest absolute
// correlation coefficient.
func CorrelationPeak(lags []int, values []float64) (lag int, value float64, err error) {
	if len(lags) != len(values) || len(values) == 0 {
		return 0, 0, fmt.Errorf("lags and correlation values must have the same non-zero length")
	}
	best := 0
	for index := 1; index < len(values); index++ {
		if math.Abs(values[index]) > math.Abs(values[best]) {
			best = index
		}
	}
	return lags[best], values[best], nil
}

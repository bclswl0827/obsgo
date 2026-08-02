package trigger

import (
	"fmt"
	"math"
)

// AICSimple calculates Maeda's sample-wise Akaike Information Criterion.
// The final value is repeated so the output has the same length as input.
func AICSimple(data []float64) []float64 {
	result := make([]float64, len(data))
	if len(data) <= 2 {
		return result
	}
	forward := onlineVariance{count: 1, mean: data[0]}
	for index := 1; index < len(data)-1; index++ {
		forward.add(data[index])
		result[index] = float64(forward.count) * math.Log(forward.variance/float64(forward.count))
	}
	backward := onlineVariance{count: 1, mean: data[len(data)-1]}
	for index := len(data) - 2; index > 0; index-- {
		backward.add(data[index])
		result[index-1] += float64(backward.count-1) * math.Log(backward.variance/float64(backward.count))
	}
	result[len(result)-1] = result[len(result)-2]
	return result
}

// AICPick returns the minimum finite AIC sample in [start, end). It is useful
// for narrowing an arrival search after a coarse trigger has found a window.
func AICPick(data []float64, start, end int) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("AIC pick requires non-empty data")
	}
	if start < 0 {
		start = 0
	}
	if end <= 0 || end > len(data) {
		end = len(data)
	}
	if start >= end {
		return 0, fmt.Errorf("invalid AIC pick interval")
	}
	values := AICSimple(data)
	best := -1
	for index := start; index < end; index++ {
		if math.IsNaN(values[index]) || math.IsInf(values[index], 0) && values[index] > 0 {
			continue
		}
		if best < 0 || values[index] < values[best] {
			best = index
		}
	}
	if best < 0 {
		return 0, fmt.Errorf("AIC interval contains no finite values")
	}
	return best, nil
}

type onlineVariance struct {
	count    int
	mean     float64
	variance float64
}

func (value *onlineVariance) add(sample float64) {
	value.count++
	delta := sample - value.mean
	value.mean += delta / float64(value.count)
	value.variance += delta * (sample - value.mean)
}

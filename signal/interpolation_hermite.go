package signal

import "math"

func centralSlopes(data []float64) []float64 {
	slopes := make([]float64, len(data))
	if len(data) < 2 {
		return slopes
	}
	slopes[0] = data[1] - data[0]
	for index := 1; index < len(data)-1; index++ {
		slopes[index] = 0.5 * (data[index+1] - data[index-1])
	}
	slopes[len(data)-1] = data[len(data)-1] - data[len(data)-2]
	return slopes
}

func weightedAverageSlopes(data []float64) []float64 {
	slopes := make([]float64, len(data))
	if len(data) < 2 {
		return slopes
	}
	m := make([]float64, len(data)-1)
	weights := make([]float64, len(m))
	for index := range m {
		m[index] = data[index+1] - data[index]
		weights[index] = 1 / math.Max(math.Abs(m[index]), math.Nextafter(1, 2)-1)
	}
	slopes[0] = m[0]
	for index := 1; index < len(data)-1; index++ {
		slopes[index] = (weights[index-1]*m[index-1] + weights[index]*m[index]) / (weights[index-1] + weights[index])
		if signOfSlope(m[index-1]) != signOfSlope(m[index]) {
			slopes[index] = 0
		}
	}
	slopes[len(data)-1] = m[len(m)-1]
	return slopes
}

func signOfSlope(value float64) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}

func interpolateHermite(data, slopes []float64, position float64) float64 {
	if position >= float64(len(data)-1) {
		return data[len(data)-1]
	}
	left := int(math.Floor(position))
	t := position - float64(left)
	t2, t3 := t*t, t*t*t
	h00 := 2*t3 - 3*t2 + 1
	h10 := t3 - 2*t2 + t
	h01 := -2*t3 + 3*t2
	h11 := t3 - t2
	return h00*data[left] + h10*slopes[left] + h01*data[left+1] + h11*slopes[left+1]
}

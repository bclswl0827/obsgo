package eewgo

import "math"

func numpyCumsumFloat(a []float64, hook func(float64) float64) []float64 {
	n := len(a)
	if n == 0 {
		return []float64{}
	}

	result := make([]float64, n)
	result[0] = hook(a[0])

	for i := 1; i < n; i++ {
		result[i] = result[i-1] + hook(a[i])
	}

	return result
}

func numpyWhereFloat(arr []float64, threshold func(float64) bool) []int {
	var indices []int
	for i, v := range arr {
		if threshold(v) {
			indices = append(indices, i)
		}
	}

	return indices
}

func numpyWhereInt(arr []int, threshold func(int) bool) []int {
	var indices []int
	for i, v := range arr {
		if threshold(v) {
			indices = append(indices, i)
		}
	}

	return indices
}

func numpyEmptyLikeBool(a []int) []bool {
	dataArr := make([]bool, len(a))
	return dataArr
}

func numpyDiffInt(a []int) []int {
	if len(a) < 2 {
		return []int{}
	}

	diff := make([]int, len(a)-1)
	for i := 1; i < len(a); i++ {
		diff[i-1] = a[i] - a[i-1]
	}

	return diff
}

func numpyZerosFloat(n int) []float64 {
	return make([]float64, n)
}

func numpySquareFloat(a []float64) []float64 {
	result := make([]float64, len(a))
	for i, v := range a {
		result[i] = v * v
	}
	return result
}

func numpyConcatenateFloat(a, b []float64) []float64 {
	return append(a, b...)
}

func numpyMeanFloat(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	var sum float64
	for _, value := range arr {
		sum += value
	}
	return sum / float64(len(arr))
}

func numpyStdFloat(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	mean := numpyMeanFloat(arr)
	var varianceSum float64
	for _, value := range arr {
		varianceSum += (value - mean) * (value - mean)
	}
	variance := varianceSum / float64(len(arr))
	return math.Sqrt(variance)
}

package trigger

import (
	"fmt"
	"math"
)

// EnergyRatio returns a short-window to long-window mean-square ratio. The
// initial long-window samples are zeroed because no stable LTA exists yet.
func EnergyRatio(data []float64, nsta, nlta int) ([]float64, error) {
	if nsta <= 0 || nlta <= 0 {
		return nil, fmt.Errorf("STA and LTA lengths must be positive")
	}
	if len(data) == 0 {
		return []float64{}, nil
	}
	squares := make([]float64, len(data))
	for index, value := range data {
		squares[index] = value * value
	}
	prefix := make([]float64, len(data)+1)
	for index, value := range squares {
		prefix[index+1] = prefix[index] + value
	}
	result := make([]float64, len(data))
	for index := nlta - 1; index < len(data); index++ {
		lta := (prefix[index+1] - prefix[index+1-nlta]) / float64(nlta)
		staStart := index + 1 - nsta
		if staStart < 0 {
			staStart = 0
		}
		sta := (prefix[index+1] - prefix[staStart]) / float64(index+1-staStart)
		if lta > 0 {
			result[index] = sta / lta
		}
	}
	return result, nil
}

// KurtosisCharacteristic computes excess kurtosis over a trailing window.
// It is useful for impulsive arrivals and is zero until a complete window is
// available.
func KurtosisCharacteristic(data []float64, window int) ([]float64, error) {
	if window <= 1 {
		return nil, fmt.Errorf("kurtosis window must contain at least two samples")
	}
	result := make([]float64, len(data))
	for index := window - 1; index < len(data); index++ {
		mean := 0.0
		for cursor := index - window + 1; cursor <= index; cursor++ {
			mean += data[cursor]
		}
		mean /= float64(window)
		second, fourth := 0.0, 0.0
		for cursor := index - window + 1; cursor <= index; cursor++ {
			deviation := data[cursor] - mean
			square := deviation * deviation
			second += square
			fourth += square * square
		}
		second /= float64(window)
		fourth /= float64(window)
		if second > 0 {
			result[index] = fourth/(second*second) - 3
		}
	}
	return result, nil
}

// ARCoefficients estimates autoregressive coefficients using ordinary least
// squares. The model is x[t] = c[0]x[t-1] + ... + c[p-1]x[t-p].
func ARCoefficients(data []float64, order int) ([]float64, error) {
	if order <= 0 || len(data) <= order {
		return nil, fmt.Errorf("AR order must be positive and shorter than data")
	}
	matrix := make([][]float64, order)
	right := make([]float64, order)
	for row := range matrix {
		matrix[row] = make([]float64, order)
	}
	for index := order; index < len(data); index++ {
		for row := 0; row < order; row++ {
			left := data[index-row-1]
			right[row] += left * data[index]
			for column := 0; column < order; column++ {
				matrix[row][column] += left * data[index-column-1]
			}
		}
	}
	return solveLinearSystem(matrix, right)
}

// ARResidual returns absolute one-step prediction errors for an AR model.
func ARResidual(data, coefficients []float64) []float64 {
	result := make([]float64, len(data))
	for index := len(coefficients); index < len(data); index++ {
		prediction := 0.0
		for order, coefficient := range coefficients {
			prediction += coefficient * data[index-order-1]
		}
		result[index] = math.Abs(data[index] - prediction)
	}
	return result
}

// ARPick returns P and optionally S arrival times in seconds using the
// ObsPy/ Akazawa AR-AIC and STA/LTA workflow. varianceP and varianceS are the
// P/S variance-window lengths in seconds, corresponding to ObsPy l_p/l_s.
func ARPick(vertical, north, east []float64, sampleRate, lowFrequency, highFrequency, ltaP, staP, ltaS, staS float64, orderP, orderS int, varianceP, varianceS float64, pickS bool) (pTime, sTime float64, err error) {
	if len(vertical) != len(north) || len(vertical) != len(east) {
		return 0, 0, fmt.Errorf("all three components must have the same length")
	}
	if !finitePositiveTrigger(sampleRate) {
		return 0, 0, fmt.Errorf("sample rate must be finite and positive")
	}
	if !finitePositiveTrigger(lowFrequency) || !finitePositiveTrigger(highFrequency) || lowFrequency >= highFrequency || highFrequency >= sampleRate/2 {
		return 0, 0, fmt.Errorf("AR pick filter frequencies must satisfy 0 < low < high < Nyquist")
	}
	if !finitePositiveTrigger(ltaP) || !finitePositiveTrigger(staP) || !finitePositiveTrigger(ltaS) || !finitePositiveTrigger(staS) {
		return 0, 0, fmt.Errorf("AR pick windows must be finite and positive")
	}
	if staP > ltaP || staS > ltaS {
		return 0, 0, fmt.Errorf("AR pick STA windows must not exceed LTA windows")
	}
	if orderP <= 0 || orderS <= 0 {
		return 0, 0, fmt.Errorf("AR model orders must be positive")
	}
	if !finitePositiveTrigger(varianceP) || !finitePositiveTrigger(varianceS) {
		return 0, 0, fmt.Errorf("AR variance windows must be finite and positive")
	}
	return arPickObsPy(vertical, north, east, sampleRate, lowFrequency, highFrequency,
		ltaP, staP, ltaS, staS, orderP, orderS, varianceP, varianceS, pickS)
}

func finitePositiveTrigger(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func maxIntTrigger(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func solveLinearSystem(matrix [][]float64, right []float64) ([]float64, error) {
	n := len(right)
	for pivot := 0; pivot < n; pivot++ {
		best := pivot
		for row := pivot + 1; row < n; row++ {
			if math.Abs(matrix[row][pivot]) > math.Abs(matrix[best][pivot]) {
				best = row
			}
		}
		if math.Abs(matrix[best][pivot]) < 1e-15 {
			return nil, fmt.Errorf("singular AR normal matrix")
		}
		matrix[pivot], matrix[best] = matrix[best], matrix[pivot]
		right[pivot], right[best] = right[best], right[pivot]
		for row := pivot + 1; row < n; row++ {
			factor := matrix[row][pivot] / matrix[pivot][pivot]
			for column := pivot; column < n; column++ {
				matrix[row][column] -= factor * matrix[pivot][column]
			}
			right[row] -= factor * right[pivot]
		}
	}
	result := make([]float64, n)
	for row := n - 1; row >= 0; row-- {
		value := right[row]
		for column := row + 1; column < n; column++ {
			value -= matrix[row][column] * result[column]
		}
		result[row] = value / matrix[row][row]
	}
	return result, nil
}

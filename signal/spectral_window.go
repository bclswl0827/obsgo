package signal

import (
	"fmt"
	"math"
)

func spectralWindow(window SpectralWindow, length int) ([]float64, error) {
	if length <= 0 {
		return nil, fmt.Errorf("window length must be positive")
	}
	result := make([]float64, length)
	if window == DefaultSpectralWindow {
		window = HannWindow
	}
	for index := range result {
		phase := 2 * math.Pi * float64(index) / float64(length)
		switch window {
		case RectangularWindow:
			result[index] = 1
		case HannWindow:
			result[index] = 0.5 * (1 - math.Cos(phase))
		case HammingWindow:
			result[index] = 0.54 - 0.46*math.Cos(phase)
		case WelchWindow:
			result = obspyWelchWindow(length)
			return result, nil
		default:
			return nil, fmt.Errorf("unsupported spectral window %d", window)
		}
	}
	return result, nil
}

func obsPySpectralWindow(window SpectralWindow, length int) ([]float64, error) {
	if length <= 0 {
		return nil, fmt.Errorf("window length must be positive")
	}
	if window == DefaultSpectralWindow {
		window = HannWindow
	}
	if window == RectangularWindow || window == WelchWindow {
		return spectralWindow(window, length)
	}
	result := make([]float64, length)
	if length == 1 {
		result[0] = 1
		return result, nil
	}
	for index := range result {
		phase := 2 * math.Pi * float64(index) / float64(length-1)
		switch window {
		case HannWindow:
			result[index] = 0.5 * (1 - math.Cos(phase))
		case HammingWindow:
			result[index] = 0.54 - 0.46*math.Cos(phase)
		default:
			return nil, fmt.Errorf("unsupported spectral window %d", window)
		}
	}
	return result, nil
}

func obspyWelchWindow(length int) []float64 {
	if length <= 2 {
		return make([]float64, length)
	}
	half := (length + 1) / 2
	left := make([]float64, half)
	for index := range left {
		x := float64(index) / float64(half)
		left[index] = 1 - x*x
	}
	var right []float64
	if half%2 == 0 {
		right = append([]float64(nil), left...)
	} else {
		right = make([]float64, half-1)
		for index := range right {
			x := float64(index) / float64(half-1)
			right[index] = 1 - x*x
		}
	}
	reverseFloat64(left)
	left[0] = 0
	right[len(right)-1] = 0
	return append(left, right...)
}

func reverseFloat64(values []float64) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

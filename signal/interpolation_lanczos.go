package signal

import (
	"math"
)

func interpolateLanczos(data []float64, position float64, width int, window LanczosKernelWindow) float64 {
	base := int(math.Floor(position))
	result := 0.0
	for index := base - width; index <= base+width; index++ {
		if index < 0 || index >= len(data) {
			continue
		}
		distance := position - float64(index)
		if math.Abs(distance) > float64(width) {
			continue
		}
		result += data[index] * lanczosKernel(distance, width, window)
	}
	return result
}

func lanczosKernel(distance float64, width int, window LanczosKernelWindow) float64 {
	if math.Abs(distance) > float64(width) {
		return 0
	}
	sinc := sincValue(distance)
	var taper float64
	switch window {
	case LanczosWindow:
		taper = sincValue(distance / float64(width))
	case HanningLanczosWindow:
		taper = 0.5 * (1 + math.Cos(math.Pi*distance/float64(width)))
	case BlackmanLanczosWindow:
		taper = 0.42 + 0.5*math.Cos(math.Pi*distance/float64(width)) + 0.08*math.Cos(2*math.Pi*distance/float64(width))
	}
	return sinc * taper
}

func sincValue(value float64) float64 {
	if math.Abs(value) < 1e-10 {
		return 1
	}
	return math.Sin(math.Pi*value) / (math.Pi * value)
}

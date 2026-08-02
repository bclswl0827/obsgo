package signal

import (
	"fmt"
	"math"
)

// CosineTaper applies ObsPy's obspy.signal.invsim.cosine_taper semantics.
// percentage is the total tapered fraction across both ends.
func CosineTaper(data []float64, percentage float64, halfCosine, sacTaper bool) ([]float64, error) {
	window, err := cosineTaperWindow(len(data), percentage, halfCosine, sacTaper)
	if err != nil {
		return nil, err
	}
	output := make([]float64, len(data))
	for index, value := range data {
		output[index] = value * window[index]
	}
	return output, nil
}

// TraceCosineTaper matches Trace.taper(type="cosine") where maxPercentage
// specifies the maximum tapered fraction at each end.
func TraceCosineTaper(data []float64, maxPercentage float64) ([]float64, error) {
	if math.IsNaN(maxPercentage) || math.IsInf(maxPercentage, 0) || maxPercentage < 0 || maxPercentage > 0.5 {
		return nil, fmt.Errorf("maximum taper percentage must be between 0 and 0.5")
	}
	length := len(data)
	windowLength := int(maxPercentage * float64(length))
	if windowLength > length/2 {
		windowLength = length / 2
	}
	if windowLength == 0 {
		return append([]float64(nil), data...), nil
	}
	sideLength := 2*windowLength + 1
	if 2*windowLength == length {
		sideLength = length
	}
	sides, err := cosineTaperWindow(sideLength, 1, true, false)
	if err != nil {
		return nil, err
	}
	output := append([]float64(nil), data...)
	for index := 0; index < windowLength; index++ {
		output[index] *= sides[index]
		output[length-windowLength+index] *= sides[len(sides)-windowLength+index]
	}
	return output, nil
}

func cosineTaperWindow(length int, percentage float64, halfCosine, sacTaper bool) ([]float64, error) {
	if math.IsNaN(percentage) || math.IsInf(percentage, 0) || percentage < 0 || percentage > 1 {
		return nil, fmt.Errorf("taper percentage must be between 0 and 1")
	}
	if length == 0 {
		return []float64{}, nil
	}
	fraction := 0
	if percentage == 0 || percentage == 1 {
		fraction = int(float64(length) * percentage / 2)
	} else {
		fraction = int(float64(length)*percentage/2 + 0.5)
	}
	leftStart := 0
	leftEnd := fraction - 1
	rightStart := length - fraction
	rightEnd := length - 1
	if sacTaper {
		leftEnd++
		rightStart--
	}
	if leftStart == leftEnd {
		leftEnd++
	}
	if rightStart == rightEnd {
		rightStart--
	}

	window := make([]float64, length)
	for index := leftStart; index <= leftEnd && index < length; index++ {
		if index < 0 {
			continue
		}
		if halfCosine {
			window[index] = 0.5 * (1 - math.Cos(math.Pi*float64(index-leftStart)/float64(leftEnd-leftStart)))
		} else {
			window[index] = math.Cos(-math.Pi / 2 * float64(leftEnd-index) / float64(leftEnd-leftStart))
		}
	}
	for index := leftEnd + 1; index < rightStart && index < length; index++ {
		if index >= 0 {
			window[index] = 1
		}
	}
	for index := max(0, rightStart); index <= rightEnd; index++ {
		if halfCosine {
			window[index] = 0.5 * (1 + math.Cos(math.Pi*float64(rightStart-index)/float64(rightEnd-rightStart)))
		} else {
			window[index] = math.Cos(math.Pi / 2 * float64(rightStart-index) / float64(rightEnd-rightStart))
		}
	}
	return window, nil
}

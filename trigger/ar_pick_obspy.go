package trigger

import (
	"fmt"
	"math"

	"github.com/bclswl0827/obsgo/signal"
)

func arPickObsPy(vertical, north, east []float64, sampleRate, lowFrequency, highFrequency, ltaP, staP, ltaS, staS float64, orderP, orderS int, varianceP, varianceS float64, pickS bool) (float64, float64, error) {
	if len(vertical) < 4 {
		return 0, 0, fmt.Errorf("AR pick requires at least four samples")
	}
	if varianceP <= 0 || varianceS <= 0 {
		return 0, 0, fmt.Errorf("P and S variance windows must be positive")
	}
	prepared := [3][]float32{}
	for component, data := range [][]float64{vertical, north, east} {
		detrended, err := signal.RemoveTrend(data, sampleRate)
		if err != nil {
			return 0, 0, err
		}
		prepared[component] = float32Samples(detrended)
	}
	scaleARPickFloat32(prepared[0], prepared[1], prepared[2])
	return arPickerFloat32(prepared[0], prepared[1], prepared[2], float32(sampleRate), float32(lowFrequency), float32(highFrequency),
		float32(ltaP), float32(staP), float32(ltaS), float32(staS), orderP, orderS, varianceP, varianceS, pickS)
}

func float32Samples(data []float64) []float32 {
	result := make([]float32, len(data))
	for index, value := range data {
		result[index] = float32(value)
	}
	return result
}

func scaleARPickFloat32(vertical, north, east []float32) {
	maximum := maxAbsFloat32(vertical)
	if maximum < 100 && maximum > 0 {
		for index := range vertical {
			vertical[index] *= 1e6
			vertical[index] /= maximum
		}
	}
	maximum = maxFloat32(maxAbsFloat32(north), maxAbsFloat32(east))
	if maximum < 100 && maximum > 0 {
		for index := range north {
			north[index] *= 1e6
			north[index] /= maximum
			east[index] *= 1e6
			east[index] /= maximum
		}
	}
}

func arPickerFloat32(vertical, north, east []float32, sampleRate, lowFrequency, highFrequency, ltaP, staP, ltaS, staS float32, orderP, orderS int, varianceP, varianceS float64, pickS bool) (float64, float64, error) {
	length := len(vertical)
	buff1 := append([]float32(nil), vertical...)
	buff1S := append([]float32(nil), north...)
	buff2 := make([]float32, length)
	buff3 := make([]float32, length)
	buff4 := make([]float32, length)
	buff4S := append([]float32(nil), east...)
	forwardError := make([]float32, length)
	backwardError := make([]float32, length)
	staBuffer := make([]float32, length)
	ltaBuffer := make([]float32, length)

	arBandpassFloat32(buff1, 1/float64(sampleRate), float64(lowFrequency), float64(highFrequency), 2, true)
	copy(buff4, buff1)
	arBandpassFloat32(buff1S, 1/float64(sampleRate), float64(lowFrequency), float64(highFrequency), 2, true)
	arBandpassFloat32(buff4S, 1/float64(sampleRate), float64(lowFrequency), float64(highFrequency), 2, true)
	selected := buff1S
	maximumHorizontal := float32(0)
	for index := 0; index < length; index++ {
		if value := absFloat32(buff1S[index]); value > maximumHorizontal {
			maximumHorizontal = value
			selected = buff1S
		}
		if value := absFloat32(buff4S[index]); value > maximumHorizontal {
			maximumHorizontal = value
			selected = buff4S
		}
	}
	copy(buff4S, selected)

	maximumVertical := maxAbsFloat32(buff4)
	envelopeMaximum := float32(0)
	i1 := 0
	for index := 0; index < length; index++ {
		buff2[index] = absFloat32(buff4[index])/maximumVertical - buff4[index]*buff4[index]/(maximumVertical*maximumVertical)
		if index == 0 || buff2[index] > buff2[index-1] {
			buff3[index] = buff2[index]
		} else {
			buff3[index] = buff2[index-1]
		}
		if buff3[index] > envelopeMaximum {
			envelopeMaximum = buff3[index]
			i1 = index
		}
	}

	nstaP := int(staP * sampleRate)
	nltaP := int(ltaP * sampleRate)
	if nstaP <= 0 || nltaP <= 0 || nstaP > nltaP {
		return 0, 0, fmt.Errorf("invalid P STA/LTA windows")
	}
	i2 := 0
	staLTAMaximum := float32(0)
	for index := 0; index < i1-nltaP; index++ {
		for sample := index + nltaP - nstaP; sample < index+nltaP; sample++ {
			staBuffer[index+nltaP] += buff3[sample] / float32(nstaP)
		}
		for sample := index; sample < index+nltaP; sample++ {
			ltaBuffer[index+nltaP] += buff3[sample] / float32(nltaP)
		}
		if ltaBuffer[index+nltaP] > 0 {
			ratio := staBuffer[index+nltaP] / ltaBuffer[index+nltaP]
			if ratio > staLTAMaximum {
				staLTAMaximum = ratio
				i2 = index + nltaP
			}
		}
	}

	nlP := int(varianceP * float64(sampleRate))
	nlS := int(varianceS * float64(sampleRate))
	if nlP <= orderP || nlS <= orderS {
		return 0, 0, fmt.Errorf("AR variance windows must exceed their model orders")
	}
	i2 += orderP + nlP
	if i2+nlP+orderP > length {
		i2 = length - nlP - orderP
	}
	if i2 <= 0 || i2+nlP > length {
		return 0, 0, fmt.Errorf("P AR search interval is empty")
	}

	for index := 0; index < i2+nlP; index++ {
		buff1[index] *= buff1[index] * buff1[index]
	}
	for index := 0; index < i2+nlP; index++ {
		buff2[i2+nlP-index-1] = buff1[index]
	}
	forwardCoefficients, err := arCoefficientsPAZ(buff1[:nlP], orderP)
	if err != nil {
		return 0, 0, err
	}
	backwardCoefficients, err := arCoefficientsPAZ(buff2[:nlP], orderP)
	if err != nil {
		return 0, 0, err
	}
	arPredictionErrors(buff1, forwardCoefficients, nlP, orderP, i2, forwardError)
	arPredictionErrors(buff2, backwardCoefficients, nlP, orderP, i2, backwardError)
	i3 := jointAICMinimum(forwardError, backwardError, orderP+nlP, i2-1-orderP-nlP, i2)

	n32 := int(math.Min(float64(length), float64((i2-i3)*2+orderP+nlP)+0.5))
	i3 = i2 - n32
	if i3 < 0 {
		i3 = 0
	}
	for index, source := 0, i3; index < n32 && source < i2; index, source = index+1, source+1 {
		buff1[index] = vertical[source] * vertical[source] * vertical[source]
	}
	for index := 0; index < n32; index++ {
		buff2[n32-index-1] = buff1[index]
	}
	clear(forwardError)
	clear(backwardError)
	forwardCoefficients, err = arCoefficientsPAZ(buff1[:nlP], orderP)
	if err != nil {
		return 0, 0, err
	}
	backwardCoefficients, err = arCoefficientsPAZ(buff2[:nlP], orderP)
	if err != nil {
		return 0, 0, err
	}
	arPredictionErrors(buff1[:n32], forwardCoefficients, nlP, orderP, n32, forwardError)
	arPredictionErrors(buff2[:n32], backwardCoefficients, nlP, orderP, n32, backwardError)
	i4Local := jointAICMinimum(forwardError, backwardError, orderP+nlP, n32-1-orderP-nlP, n32)
	i4 := 0
	if i4Local > 0 {
		i4 = i4Local + i3
	}
	pTime := float64(float32(i4-nlP) / sampleRate)
	if !pickS {
		return pTime, 0, nil
	}

	clear(forwardError)
	clear(backwardError)
	clear(staBuffer)
	clear(ltaBuffer)
	nstaS := int(staS * sampleRate)
	nltaS := int(ltaS * sampleRate)
	if nstaS <= 0 || nltaS <= 0 || nstaS > nltaS {
		return 0, 0, fmt.Errorf("invalid S STA/LTA windows")
	}
	for index := 0; index < length-nltaS; index++ {
		for sample := index + nltaS - nstaS; sample < index+nltaS; sample++ {
			staBuffer[index+nltaS] += absFloat32(buff4S[sample]) / float32(nstaS)
		}
		for sample := index; sample < index+nltaS; sample++ {
			ltaBuffer[index+nltaS] += absFloat32(buff4S[sample]) / float32(nltaS)
		}
	}
	i5 := 0
	ltaMaximum := float32(0)
	for index := i4; index < length-nltaS; index++ {
		difference := staBuffer[index+nltaS] - ltaBuffer[index+nltaS]
		if difference > ltaMaximum {
			ltaMaximum = difference
			i5 = index + nltaS
		}
	}
	i5 += orderS + nlS
	if i5 > length {
		i5 = length
	}
	clear(staBuffer)
	clear(ltaBuffer)
	for index := length - nltaS - 1; index > nltaS; index-- {
		for sample := index + nstaS - 1; sample >= index; sample-- {
			staBuffer[index] += absFloat32(buff4S[sample]) / float32(nstaS)
		}
		for sample := index + nltaS - 1; sample >= index; sample-- {
			ltaBuffer[index] += absFloat32(buff4S[sample]) / float32(nltaS)
		}
	}
	i6 := 0
	ltaMaximum = 0
	for index := i5 + nltaS; index >= i4; index-- {
		if index < i5 && index-nltaS >= 0 {
			difference := staBuffer[index-nltaS] - ltaBuffer[index-nltaS]
			if difference < ltaMaximum {
				ltaMaximum = difference
				i6 = index - nltaS
			}
		}
	}
	if i6 <= 0 {
		return pTime, 0, nil
	}
	n65 := i5 - i6
	if n65 > length || n65 <= 0 {
		return pTime, 0, nil
	}
	for index, source := 0, i6; index < n65 && source < i5; index, source = index+1, source+1 {
		buff1S[index] = buff4S[source] * buff4S[source] * buff4S[source]
	}
	for index := 0; index < n65; index++ {
		buff2[n65-index-1] = buff1S[index]
	}
	forwardCoefficients, err = arCoefficientsPAZ(buff1S[:nlS], orderS)
	if err != nil {
		return 0, 0, err
	}
	backwardCoefficients, err = arCoefficientsPAZ(buff2[:nlS], orderS)
	if err != nil {
		return 0, 0, err
	}
	arPredictionErrors(buff1S[:n65], forwardCoefficients, nlS, orderS, n65, forwardError)
	arPredictionErrors(buff2[:n65], backwardCoefficients, nlS, orderS, n65, backwardError)
	i7Local := jointAICMinimum(forwardError, backwardError, orderS+nlS, n65-1-orderS-nlS, n65)
	if i7Local <= 0 {
		return pTime, 0, nil
	}
	sTime := float64(float32(i7Local+i6-nlS) / sampleRate)
	return pTime, sTime, nil
}

func arPredictionErrors(data, coefficients []float32, window, order, limit int, output []float32) {
	for index := order; index < limit-window; index++ {
		for offset := 0; offset < window; offset++ {
			predictionError := data[index+offset]
			for coefficient := 0; coefficient < order; coefficient++ {
				predictionError -= data[index+offset-coefficient] * coefficients[coefficient]
			}
			output[index+window] += predictionError * predictionError / float32(window)
		}
	}
}

func jointAICMinimum(forwardError, backwardError []float32, start, backwardStart, limit int) int {
	best := 0
	minimum := float32(0)
	for left, right := start, backwardStart; left <= backwardStart && right >= start && left < limit; left, right = left+1, right-1 {
		if forwardError[left] <= 0 || backwardError[right] <= 0 {
			continue
		}
		forward := float64(forwardError[left])
		backward := float64(backwardError[right])
		value := float32(-math.Log(forward*forward) - math.Log(backward*backward))
		if value < minimum {
			minimum = value
			best = left
		}
	}
	return best
}

func maxAbsFloat32(data []float32) float32 {
	maximum := float32(0)
	for _, value := range data {
		if absolute := absFloat32(value); absolute > maximum {
			maximum = absolute
		}
	}
	return maximum
}

func absFloat32(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
}

func maxFloat32(left, right float32) float32 {
	if left > right {
		return left
	}
	return right
}

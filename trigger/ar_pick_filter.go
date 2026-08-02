package trigger

import (
	"fmt"
	"math"
)

// arBandpassFloat32 is a direct port of ObsPy's spr_bp_fast_bworth. Its state
// is intentionally retained between the forward and reverse passes.
func arBandpassFloat32(data []float32, samplePeriod, lowFrequency, highFrequency float64, sections int, zeroPhase bool) {
	samplePeriod = float64(float32(samplePeriod))
	lowFrequency = float64(float32(lowFrequency))
	highFrequency = float64(float32(highFrequency))
	var a, b, c, d, e [11]float64
	var state [12][6]float64
	lowAngle := lowFrequency * math.Pi * samplePeriod
	highAngle := highFrequency * math.Pi * samplePeriod
	w1 := math.Sin(lowAngle) / math.Cos(lowAngle)
	w2 := math.Sin(highAngle) / math.Cos(highAngle)
	wc := w2 - w1
	q := wc*wc + 2*w1*w2
	s := w1 * w1 * w2 * w2
	for section := 1; section <= sections; section++ {
		c1 := float64(section + sections)
		c2 := float64(4 * sections)
		cs := math.Cos((2*c1 - 1) * math.Pi / c2)
		p := -2 * wc * cs
		r := p * w1 * w2
		x := 1 + p + q + r + s
		a[section] = wc * wc / x
		b[section] = (-4 - 2*p + 2*r + 4*s) / x
		c[section] = (6 - 2*q + 6*s) / x
		d[section] = (-4 + 2*p - 2*r + 4*s) / x
		e[section] = (1 - p + q - r + s) / x
	}
	apply := func(sample int) {
		state[1][5] = float64(data[sample])
		for section := 1; section <= sections; section++ {
			value := a[section] * (state[section][5] - 2*state[section][3] + state[section][1])
			value -= b[section]*state[section+1][4] + c[section]*state[section+1][3]
			state[section+1][5] = value - d[section]*state[section+1][2] - e[section]*state[section+1][1]
		}
		for section := 1; section <= sections+1; section++ {
			for previous := 1; previous <= 4; previous++ {
				state[section][previous] = state[section][previous+1]
			}
		}
		data[sample] = float32(state[sections+1][5])
	}
	for sample := range data {
		apply(sample)
	}
	if zeroPhase {
		for sample := len(data) - 1; sample >= 0; sample-- {
			apply(sample)
		}
	}
}

// arCoefficientsPAZ directly follows ObsPy's spr_coef_paz float arithmetic.
func arCoefficientsPAZ(data []float32, order int) ([]float32, error) {
	length := len(data)
	if order <= 0 || length <= order || length < 2 {
		return nil, fmt.Errorf("AR order must be positive and shorter than its variance window")
	}
	first := make([]float32, length)
	second := make([]float32, length)
	previous := make([]float32, order)
	coefficients := make([]float32, order)
	first[1] = data[0]
	second[length-1] = data[length-1]
	for index := 2; index <= length-1; index++ {
		first[index] = data[index-1]
		second[index-1] = data[index-1]
	}
	for stage := 1; stage <= order; stage++ {
		numerator := float32(0)
		denominator := float32(0)
		for index := 1; index <= length-stage; index++ {
			numerator += first[index] * second[index]
			denominator += first[index]*first[index] + second[index]*second[index]
		}
		if denominator == 0 {
			return nil, fmt.Errorf("singular AR coefficient calculation")
		}
		coefficients[stage-1] = 2 * numerator / denominator
		if stage != 1 {
			for index := 1; index <= stage-1; index++ {
				coefficients[index-1] = previous[index-1] - coefficients[stage-1]*previous[stage-index-1]
			}
		}
		if stage == order {
			return coefficients, nil
		}
		for index := 1; index <= stage; index++ {
			previous[index-1] = coefficients[index-1]
		}
		reflection := previous[stage-1]
		for index := 1; index <= length-stage-1; index++ {
			first[index] -= reflection * second[index]
			second[index] = second[index+1] - reflection*first[index+1]
		}
	}
	return coefficients, nil
}

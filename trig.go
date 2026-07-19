package eewgo

import (
	"math"
)

// numpy.finfo(float64).tiny is the smallest positive *normal* float64. It is
// deliberately not math.SmallestNonzeroFloat64, which is a subnormal value.
const float64Tiny = 0x1p-1022

// Computes the standard STA/LTA from a given input array a. The length of
// the STA is given by nsta in samples, respectively is the length of the
// LTA given by nlta in samples. Translated from Obspy.
// https://docs.obspy.org/tutorial/code_snippets/trigger_tutorial.html#classic-sta-lta
func ClassicStaLta(a []float64, nsta, nlta int) []float64 {
	// The cumulative sum can be exploited to calculate a moving average
	sta := numpyCumsumFloat(a, func(f float64) float64 { return f * f })
	lta := make([]float64, len(sta))
	copy(lta, sta)

	// Compute for STA
	if nsta < len(sta) {
		tempArrLen := len(sta) - nsta
		tempArr1 := append([]float64(nil), sta[:tempArrLen]...)
		tempArr2 := append([]float64(nil), sta[nsta:]...)
		for i := 0; i < tempArrLen; i++ {
			sta[i+nsta] = tempArr2[i] - tempArr1[i]
		}
	}
	for i := range lta {
		sta[i] /= float64(nsta)
	}

	// Compute for LTA
	if nlta < len(lta) {
		tempArrLen := len(lta) - nlta
		tempArr1 := append([]float64(nil), lta[:tempArrLen]...)
		tempArr2 := append([]float64(nil), lta[nlta:]...)
		for i := 0; i < tempArrLen; i++ {
			lta[i+nlta] = tempArr2[i] - tempArr1[i]
		}
	}
	for i := range lta {
		lta[i] /= float64(nlta)
	}

	// Pad zeros
	for i := 0; i < min(nlta-1, len(sta)); i++ {
		sta[i] = 0
	}

	// Avoid division by zero by setting zero values to tiny float
	dtiny := float64Tiny
	for i, val := range lta {
		if val < dtiny {
			lta[i] = dtiny
		}
	}

	result := make([]float64, len(sta))
	for i := range sta {
		result[i] = sta[i] / lta[i]
	}
	return result
}

// Recursive STA/LTA, translated from Obspy.
// https://docs.obspy.org/tutorial/code_snippets/trigger_tutorial.html#recursive-sta-lta
func RecursiveStaLta(a []float64, nsta, nlta int) []float64 {
	ndat := len(a)
	// compute the short time average (STA) and long time average (LTA)
	// given by Evans and Allen
	csta := 1 / float64(nsta)
	clta := 1 / float64(nlta)
	sta := 0.0
	lta := float64Tiny // avoid zero division
	a = numpySquareFloat(a)
	charfct := numpyZerosFloat(ndat)
	icsta := 1 - csta
	iclta := 1 - clta
	for i := 1; i < ndat; i++ {
		sta = csta*a[i] + icsta*float64(sta)
		lta = clta*a[i] + iclta*float64(lta)
		charfct[i] = sta / lta
	}
	for i := 0; i < min(nlta, ndat); i++ {
		charfct[i] = 0
	}

	return charfct
}

// Delayed STA/LTA. Translated from Obspy.
// https://docs.obspy.org/tutorial/code_snippets/trigger_tutorial.html#delayed-sta-lta
func DelayedStaLta(a []float64, nsta, nlta int) []float64 {
	m := len(a)

	// compute the short time average (STA) and long time average (LTA)
	// don't start for STA at nsta because it's muted later anyway
	sta := numpyZerosFloat(m)
	lta := numpyZerosFloat(m)
	for i := 0; i < m; i++ {
		index1 := i - nsta
		if index1 < 0 {
			index1 = m - nsta + i
		}
		index2 := i - 1
		if index2 < 0 {
			index2 = m - 1 + i
		}
		index3 := i - nsta - 1
		if index3 < 0 {
			index3 = m - (nsta + 1) + i
		}
		index4 := i - nsta - nlta - 1
		if index4 < 0 {
			index4 = m - (nsta + nlta + 1) + i
		}
		sta[i] = (a[i]*a[i]+a[index1]*a[index1])/float64(nsta) + sta[index2]
		lta[i] = (a[index3]*a[index3]+a[index4]*a[index4])/float64(nlta) + lta[index2]
	}

	for i := 0; i < min(nlta+nsta+50, m); i++ {
		sta[i] = 0
		lta[i] = 1 // avoid division by zero
	}

	result := make([]float64, len(sta))
	for i := range sta {
		result[i] = sta[i] / lta[i]
	}
	return result
}

// Z-detector. Translated from Obspy.
// https://docs.obspy.org/tutorial/code_snippets/trigger_tutorial.html#z-detect
func ZDetect(a []float64, nsta int) []float64 {
	// Z-detector given by Swindell and Snell (1977)
	// Standard Sta shifted by 1
	sta := numpyCumsumFloat(a, func(f float64) float64 { return f * f })

	tempArr1 := make([]float64, len(sta)-nsta-1)
	copy(tempArr1, sta[:len(sta)-nsta-1])
	tempArr2 := make([]float64, len(sta)-nsta-1)
	copy(tempArr2, sta[nsta:len(sta)-1])
	for i := 0; i < len(tempArr1); i++ {
		sta[i+nsta+1] = tempArr2[i] - tempArr1[i]
	}
	sta[nsta] = sta[nsta-1]
	for i := 0; i < nsta; i++ {
		sta[i] = 0
	}
	a_mean := numpyMeanFloat(sta)
	a_std := numpyStdFloat(sta)

	_z := make([]float64, len(sta))
	for i := range _z {
		_z[i] = (sta[i] - a_mean) / a_std
	}
	return _z
}

// Computes the carlSTAtrig characteristic function. Translated from Obspy.
// https://docs.obspy.org/tutorial/code_snippets/trigger_tutorial.html#carl-sta-trig
func CarlStaTrig(a []float64, nsta, nlta int, ratio, quiet float64) []float64 {
	m := len(a)

	sta := numpyZerosFloat(len(a))
	lta := numpyZerosFloat(len(a))
	star := numpyZerosFloat(len(a))
	ltar := numpyZerosFloat(len(a))
	pad_sta := numpyZerosFloat(nsta)
	pad_lta := numpyZerosFloat(nlta) // avoid for 0 division 0/1=0

	// compute the short time average (STA)
	for i := 0; i < nsta; i++ {
		// window size to smooth over
		n := numpyConcatenateFloat(pad_sta, a[i:m-nsta+i])
		for j := 0; j < len(sta); j++ {
			sta[j] += n[j]
		}
	}
	for i := 0; i < len(sta); i++ {
		sta[i] /= float64(nsta)
	}

	// compute the long time average (LTA), 8 sec average over sta
	for i := 0; i < nlta; i++ {
		// window size to smooth over
		n := numpyConcatenateFloat(pad_lta, sta[i:m-nlta+i])
		for j := 0; j < len(lta); j++ {
			lta[j] += n[j]
		}
	}
	for i := 0; i < len(lta); i++ {
		lta[i] /= float64(nlta)
	}
	lta = numpyConcatenateFloat(numpyZerosFloat(1), lta)[:m] // XXX ???

	// compute star, average of abs diff between trace and lta
	for i := 0; i < nsta; i++ {
		// window size to smooth over
		abs := make([]float64, len(a[i:m-nsta+i]))
		for j, k := i, 0; j < m-nsta+i; j++ {
			abs[k] = math.Abs(a[j] - lta[j])
			k++
		}
		n := numpyConcatenateFloat(pad_sta, abs)
		for j := 0; j < len(star); j++ {
			star[j] += n[j]
		}
	}
	for i := 0; i < len(star); i++ {
		star[i] /= float64(nsta)
	}

	// compute ltar, 8 sec average over star
	for i := 0; i < nlta; i++ {
		// window size to smooth over
		n := numpyConcatenateFloat(pad_lta, star[i:m-nlta+i])
		for j := 0; j < len(ltar); j++ {
			ltar[j] += n[j]
		}
	}
	for i := 0; i < len(ltar); i++ {
		ltar[i] /= float64(nlta)
	}

	eta := make([]float64, m)
	for i := 0; i < m; i++ {
		eta[i] = star[i] - (ratio * ltar[i]) - math.Abs(sta[i]-lta[i]) - quiet
	}
	for i := 0; i < nlta; i++ {
		eta[i] = -1.0
	}
	return eta
}

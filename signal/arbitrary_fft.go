package signal

import (
	"math"
	"math/cmplx"
)

// responseFFTLength follows ObsPy's _npts2nfft strategy: at least twice the
// data length, preferring nearby lengths without large prime factors.
func responseFFTLength(numPoints int) int {
	length := 2 * numPoints
	if numPoints%2 != 0 {
		length = 2 * (numPoints + 1)
	}
	if length > 5000 && largestPrimeFactor(length) >= 500 {
		for increment := 2; increment <= 20; increment += 2 {
			candidate := length + increment
			if largestPrimeFactor(candidate) < 500 {
				return candidate
			}
		}
		length, _ = FindPowerTwo(length)
	}
	return length
}

func largestPrimeFactor(value int) int {
	largest := 1
	for divisor := 2; divisor*divisor <= value; divisor++ {
		for value%divisor == 0 {
			largest = divisor
			value /= divisor
		}
	}
	if value > largest {
		largest = value
	}
	return largest
}

// fftExactInPlace supports arbitrary transform lengths through Bluestein's
// algorithm while retaining the radix-2 implementation for power-of-two data.
func fftExactInPlace(values []complex128, inverse bool) {
	if len(values) == 0 {
		return
	}
	if len(values)&(len(values)-1) == 0 {
		fftInPlace(values, inverse)
		return
	}
	if inverse {
		for index := range values {
			values[index] = cmplx.Conj(values[index])
		}
		bluesteinForward(values)
		scale := complex(float64(len(values)), 0)
		for index := range values {
			values[index] = cmplx.Conj(values[index]) / scale
		}
		return
	}
	bluesteinForward(values)
}

func bluesteinForward(values []complex128) {
	length := len(values)
	convolutionLength, _ := FindPowerTwo(2*length - 1)
	left := make([]complex128, convolutionLength)
	right := make([]complex128, convolutionLength)
	for index, value := range values {
		angle := math.Pi * float64(index) * float64(index) / float64(length)
		forwardChirp := cmplx.Rect(1, -angle)
		inverseChirp := cmplx.Rect(1, angle)
		left[index] = value * forwardChirp
		right[index] = inverseChirp
		if index != 0 {
			right[convolutionLength-index] = inverseChirp
		}
	}
	fftInPlace(left, false)
	fftInPlace(right, false)
	for index := range left {
		left[index] *= right[index]
	}
	fftInPlace(left, true)
	for index := range values {
		angle := math.Pi * float64(index) * float64(index) / float64(length)
		values[index] = left[index] * cmplx.Rect(1, -angle)
	}
}

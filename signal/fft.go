package signal

import (
	"fmt"
	"math"
	"math/cmplx"
)

// FindPowerTwo returns the smallest power of two greater than or equal to length.
func FindPowerTwo(length int) (size, log2 int) {
	if length <= 2 {
		return 2, 1
	}
	size, log2 = 2, 1
	for size < length {
		size *= 2
		log2++
	}
	return size, log2
}

// CalcDFT returns the packed real FFT representation used by SeisPlotJS:
// real values occupy indexes [0,N/2], and imag(X[k]) is stored at N-k.
func CalcDFT(data []float64) []float64 {
	size, _ := FindPowerTwo(len(data))
	if size < 16 {
		size = 16
	}
	values := make([]complex128, size)
	for index, value := range data {
		values[index] = complex(float64(float32(value)), 0)
	}
	fftInPlace(values, false)
	packed := packRealSpectrum(values)
	for index := range packed {
		packed[index] = float64(float32(packed[index]))
	}
	return packed
}

func InverseDFT(packed []float64, originalLength int) ([]float64, error) {
	if originalLength < 0 || originalLength > len(packed) {
		return nil, fmt.Errorf("original length %d exceeds packed FFT length %d", originalLength, len(packed))
	}
	if len(packed) < 16 || len(packed)&(len(packed)-1) != 0 {
		return nil, fmt.Errorf("packed FFT length must be a power of two")
	}
	values, err := unpackRealSpectrum(packed)
	if err != nil {
		return nil, err
	}
	fftInPlace(values, true)
	output := make([]float64, originalLength)
	for index := range output {
		output[index] = float64(float32(real(values[index])))
	}
	return output, nil
}

// ComplexFFT calculates an unnormalized forward transform, padding to a power of two.
func ComplexFFT(data []float64) []complex128 {
	size, _ := FindPowerTwo(len(data))
	values := make([]complex128, size)
	for index, value := range data {
		values[index] = complex(value, 0)
	}
	fftInPlace(values, false)
	return values
}

// ComplexIFFT calculates a normalized inverse transform.
func ComplexIFFT(spectrum []complex128) ([]complex128, error) {
	if len(spectrum) == 0 || len(spectrum)&(len(spectrum)-1) != 0 {
		return nil, fmt.Errorf("spectrum length must be a power of two")
	}
	output := append([]complex128(nil), spectrum...)
	fftInPlace(output, true)
	return output, nil
}

type FFTResult struct {
	OriginalLength int
	Packed         []float64
	NumPoints      int
	SampleRate     float64
	InputUnits     string
}

func FFTForward(data []float64, sampleRate float64) (*FFTResult, error) {
	return NewFFTResultFromPacked(CalcDFT(data), len(data), sampleRate)
}

func NewFFTResultFromPacked(packed []float64, originalLength int, sampleRate float64) (*FFTResult, error) {
	if originalLength < 0 || originalLength > len(packed) {
		return nil, fmt.Errorf("not enough FFT points for original length %d", originalLength)
	}
	if len(packed) == 0 || len(packed)&(len(packed)-1) != 0 {
		return nil, fmt.Errorf("packed FFT length must be a power of two")
	}
	return &FFTResult{
		OriginalLength: originalLength,
		Packed:         append([]float64(nil), packed...),
		NumPoints:      len(packed),
		SampleRate:     sampleRate,
	}, nil
}

func NewFFTResultFromComplex(values []complex128, originalLength int, sampleRate float64) (*FFTResult, error) {
	if len(values) < 2 {
		return nil, fmt.Errorf("complex spectrum requires DC and Nyquist values")
	}
	size := 2 * (len(values) - 1)
	full := make([]complex128, size)
	copy(full, values)
	for index := 1; index < len(values)-1; index++ {
		full[size-index] = cmplx.Conj(values[index])
	}
	return NewFFTResultFromPacked(packRealSpectrum(full), originalLength, sampleRate)
}

func NewFFTResultFromAmpPhase(amplitude, phase []float64, originalLength int, sampleRate float64) (*FFTResult, error) {
	if len(amplitude) != len(phase) {
		return nil, fmt.Errorf("amplitude and phase lengths differ")
	}
	values := make([]complex128, len(amplitude))
	for index := range values {
		values[index] = cmplx.Rect(amplitude[index], phase[index])
	}
	return NewFFTResultFromComplex(values, originalLength, sampleRate)
}

func (result *FFTResult) FundamentalFrequency() float64 {
	if result == nil || result.NumPoints == 0 {
		return 0
	}
	return result.SampleRate / float64(result.NumPoints)
}

func (result *FFTResult) NumFrequencies() int {
	if result == nil {
		return 0
	}
	return result.NumPoints/2 + 1
}

func (result *FFTResult) MinFrequency() float64 { return result.FundamentalFrequency() }

func (result *FFTResult) MaxFrequency() float64 {
	if result == nil {
		return 0
	}
	return result.SampleRate / 2
}

func (result *FFTResult) AsComplex() ([]complex128, error) {
	if result == nil {
		return nil, fmt.Errorf("FFT result is nil")
	}
	full, err := unpackRealSpectrum(result.Packed)
	if err != nil {
		return nil, err
	}
	return append([]complex128(nil), full[:len(full)/2+1]...), nil
}

func (result *FFTResult) AsAmpPhase() (amplitude, phase []float64, err error) {
	values, err := result.AsComplex()
	if err != nil {
		return nil, nil, err
	}
	amplitude = make([]float64, len(values))
	phase = make([]float64, len(values))
	for index, value := range values {
		amplitude[index] = cmplx.Abs(value)
		phase[index] = cmplx.Phase(value)
	}
	return amplitude, phase, nil
}

func (result *FFTResult) Frequencies() []float64 {
	output := make([]float64, result.NumFrequencies())
	step := result.FundamentalFrequency()
	for index := range output {
		output[index] = float64(index) * step
	}
	return output
}

func (result *FFTResult) Amplitudes() ([]float64, error) {
	amplitudes, _, err := result.AsAmpPhase()
	return amplitudes, err
}

func (result *FFTResult) Phases() ([]float64, error) {
	_, phases, err := result.AsAmpPhase()
	return phases, err
}

func (result *FFTResult) Inverse() ([]float64, error) {
	if result == nil {
		return nil, fmt.Errorf("FFT result is nil")
	}
	return InverseDFT(result.Packed, result.OriginalLength)
}

func (result *FFTResult) Clone() *FFTResult {
	if result == nil {
		return nil
	}
	output := *result
	output.Packed = append([]float64(nil), result.Packed...)
	return &output
}

func fftInPlace(values []complex128, inverse bool) {
	length := len(values)
	for i, j := 1, 0; i < length; i++ {
		bit := length >> 1
		for ; j&bit != 0; bit >>= 1 {
			j ^= bit
		}
		j ^= bit
		if i < j {
			values[i], values[j] = values[j], values[i]
		}
	}
	for block := 2; block <= length; block <<= 1 {
		angle := -2 * math.Pi / float64(block)
		if inverse {
			angle = -angle
		}
		step := cmplx.Rect(1, angle)
		for start := 0; start < length; start += block {
			factor := complex(1, 0)
			half := block / 2
			for offset := 0; offset < half; offset++ {
				even := values[start+offset]
				odd := values[start+offset+half] * factor
				values[start+offset] = even + odd
				values[start+offset+half] = even - odd
				factor *= step
			}
		}
	}
	if inverse {
		scale := complex(float64(length), 0)
		for index := range values {
			values[index] /= scale
		}
	}
}

func packRealSpectrum(values []complex128) []float64 {
	length := len(values)
	packed := make([]float64, length)
	packed[0] = real(values[0])
	packed[length/2] = real(values[length/2])
	for index := 1; index < length/2; index++ {
		packed[index] = real(values[index])
		packed[length-index] = imag(values[index])
	}
	return packed
}

func unpackRealSpectrum(packed []float64) ([]complex128, error) {
	length := len(packed)
	if length == 0 || length&1 != 0 {
		return nil, fmt.Errorf("packed FFT length must be positive and even")
	}
	values := make([]complex128, length)
	values[0] = complex(packed[0], 0)
	values[length/2] = complex(packed[length/2], 0)
	for index := 1; index < length/2; index++ {
		value := complex(packed[index], packed[length-index])
		values[index] = value
		values[length-index] = cmplx.Conj(value)
	}
	return values, nil
}

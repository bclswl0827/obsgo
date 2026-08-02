package signal

import (
	"fmt"
	"math"
)

const (
	defaultHilbertOrder    = 10
	defaultHilbertLowEdge  = 0.05
	defaultHilbertHighEdge = 0.95
)

// oregonDefaultHilbertCoefficients is the Float32 coefficient vector emitted
// by OregonDSP 1.3.1 for CenteredHilbertTransform(10, 0.05, 0.95). Keeping the
// canonical SeisPlotJS design frozen avoids platform-dependent Remez extrema.
var oregonDefaultHilbertCoefficients = []float64{
	1.1187021309311262e-9,
	-0.0960223451256752,
	1.3782458729849623e-8,
	-0.07441437989473343,
	-1.526658266470804e-8,
	-0.11460818350315094,
	3.654209157488708e-10,
	-0.20432430505752563,
	1.79707648761962e-9,
	-0.633985698223114,
	-1.79707648761962e-9,
	0.633985698223114,
	-1.79707648761962e-9,
	0.20432430505752563,
	1.79707648761962e-9,
	0.11460818350315094,
	3.654209157488708e-10,
	0.07441437989473343,
	-1.526658266470804e-8,
	0.0960223451256752,
	1.4901161193847656e-8,
}

// CenteredHilbertTransform is OregonDSP's odd-length, Type-III equiripple
// Hilbert transformer. The center coefficient is zero and the group delay is
// N samples.
type CenteredHilbertTransform struct {
	order        int
	lowEdge      float64
	highEdge     float64
	coefficients []float64
}

// NewCenteredHilbertTransform designs an OregonDSP-compatible centered
// Hilbert FIR filter with 2*order+1 coefficients.
func NewCenteredHilbertTransform(order int, lowEdge, highEdge float64) (*CenteredHilbertTransform, error) {
	if order < 1 {
		return nil, fmt.Errorf("Hilbert order must be positive")
	}
	if !(0 < lowEdge && lowEdge < highEdge && highEdge < 1) {
		return nil, fmt.Errorf("Hilbert edges must satisfy 0 < lowEdge < highEdge < 1")
	}
	var coefficients []float64
	if order == defaultHilbertOrder && lowEdge == defaultHilbertLowEdge && highEdge == defaultHilbertHighEdge {
		coefficients = append([]float64(nil), oregonDefaultHilbertCoefficients...)
	} else {
		var err error
		coefficients, err = designCenteredHilbert(order, lowEdge, highEdge)
		if err != nil {
			return nil, err
		}
	}
	return &CenteredHilbertTransform{
		order:        order,
		lowEdge:      lowEdge,
		highEdge:     highEdge,
		coefficients: coefficients,
	}, nil
}

// Coefficients returns a copy of the designed FIR coefficients.
func (filter *CenteredHilbertTransform) Coefficients() []float64 {
	if filter == nil {
		return nil
	}
	return append([]float64(nil), filter.coefficients...)
}

// Filter applies OregonDSP's Float32 FFT convolution. The output length is
// len(data)+len(coefficients)-1.
func (filter *CenteredHilbertTransform) Filter(data []float64) []float64 {
	if filter == nil {
		return nil
	}
	if len(data) == 0 {
		return []float64{}
	}
	length := len(data) + len(filter.coefficients) - 1
	fftLength, _ := FindPowerTwo(length)
	if fftLength < 16 {
		fftLength = 16
	}
	paddedData := make([]float64, fftLength)
	paddedCoefficients := make([]float64, fftLength)
	copy(paddedData, data)
	copy(paddedCoefficients, filter.coefficients)
	dataSpectrum := CalcDFT(paddedData)
	coefficientSpectrum := CalcDFT(paddedCoefficients)
	packed := oregonDFTProduct(coefficientSpectrum, dataSpectrum)
	output, _ := InverseDFT(packed, length)
	return output
}

// oregonDFTProduct preserves RDFT.dftProduct's Float32 operation order. The
// imaginary part of bin k is stored at packed[n-k].
func oregonDFTProduct(kernel, transform []float64) []float64 {
	if len(kernel) != len(transform) || len(kernel) == 0 {
		return nil
	}
	length := len(kernel)
	half := length / 2
	result := make([]float64, length)
	result[0] = float64(float32(kernel[0]) * float32(transform[0]))
	result[half] = float64(float32(kernel[half]) * float32(transform[half]))
	for index := 1; index < half; index++ {
		imaginaryIndex := length - index
		kernelReal := float32(kernel[index])
		kernelImaginary := float32(kernel[imaginaryIndex])
		transformReal := float32(transform[index])
		transformImaginary := float32(transform[imaginaryIndex])
		result[index] = float64(kernelReal*transformReal - kernelImaginary*transformImaginary)
		result[imaginaryIndex] = float64(kernelReal*transformImaginary + kernelImaginary*transformReal)
	}
	return result
}

// CenteredHilbert calculates the default OregonDSP centered Hilbert transform.
// It returns the full FIR convolution, which is len(data)+20 samples for the
// default 21-tap filter.
func CenteredHilbert(data []float64) []float64 {
	filter, err := NewCenteredHilbertTransform(
		defaultHilbertOrder,
		defaultHilbertLowEdge,
		defaultHilbertHighEdge,
	)
	if err != nil {
		return nil
	}
	return filter.Filter(data)
}

// CenteredHilbertWithParameters calculates a centered Hilbert transform with
// explicit OregonDSP design parameters.
func CenteredHilbertWithParameters(data []float64, order int, lowEdge, highEdge float64) ([]float64, error) {
	filter, err := NewCenteredHilbertTransform(order, lowEdge, highEdge)
	if err != nil {
		return nil, err
	}
	return filter.Filter(data), nil
}

func designCenteredHilbert(order int, lowEdge, highEdge float64) ([]float64, error) {
	grid := createHilbertGrid(order, lowEdge, highEdge)
	populateHilbertGrid(grid)
	if err := remezHilbert(grid); err != nil {
		return nil, err
	}
	cosineCoefficients := calculateHilbertCosineCoefficients(grid, order)
	return interpretHilbertCoefficients(cosineCoefficients, order), nil
}

type hilbertRemezGrid struct {
	grid         []float64
	x            []float64
	desired      []float64
	weight       []float64
	extrema      []int
	containsZero bool
	containsPi   bool
}

func createHilbertGrid(order int, lowEdge, highEdge float64) *hilbertRemezGrid {
	// OregonDSP allocates N+1 extrema over each band and uses a dense grid
	// with GRIDDENSITY=20. CenteredHilbertTransform has one band, so the
	// initial extrema are deterministic band-edge/equidistant points here.
	nextrema := order + 1
	gridSize := 1 + (nextrema-1)*20
	grid := make([]float64, gridSize)
	extrema := make([]int, nextrema)
	bandWidth := highEdge - lowEdge
	for i := 0; i < gridSize; i++ {
		grid[i] = lowEdge + bandWidth*float64(i)/float64(gridSize-1)
		if i%20 == 0 {
			extrema[i/20] = i
		}
	}
	x := make([]float64, gridSize)
	for i, value := range grid {
		x[i] = math.Cos(math.Pi * value)
	}
	return &hilbertRemezGrid{
		grid:         grid,
		x:            x,
		desired:      make([]float64, gridSize),
		weight:       make([]float64, gridSize),
		extrema:      extrema,
		containsZero: false,
		containsPi:   false,
	}
}

func populateHilbertGrid(grid *hilbertRemezGrid) {
	for i, value := range grid.grid {
		sine := math.Sin(math.Pi * value)
		grid.desired[i] = 1 / sine
		grid.weight[i] = sine
	}
}

func remezHilbert(grid *hilbertRemezGrid) error {
	const maxIterations = 25
	nextrema := len(grid.extrema)
	errValues := make([]float64, len(grid.grid))
	approximation := make([]float64, len(grid.grid))
	for iteration := 0; iteration < maxIterations; iteration++ {
		delta, err := hilbertRemezDelta(grid)
		if err != nil {
			return err
		}
		polynomial, err := hilbertInterpolatingPolynomial(grid, delta)
		if err != nil {
			return err
		}
		for i := range grid.grid {
			approximation[i] = evaluateBarycentric(polynomial.nodes, polynomial.values, polynomial.weights, grid.x[i])
			errValues[i] = approximation[i] - grid.desired[i]
		}

		newExtrema := make([]int, 0, nextrema)
		changes := 0
		for current := 0; current < nextrema; current++ {
			currentIndex := grid.extrema[current]
			sign := signOf(errValues[currentIndex])
			index := currentIndex + 1
			if index < len(grid.grid) {
				for index < len(grid.grid) && signOf(errValues[index]-errValues[index-1]) == sign {
					index++
				}
			}
			index--
			if index == currentIndex {
				index = currentIndex - 1
				if index >= 0 {
					for index >= 0 && signOf(errValues[index]-errValues[index+1]) == sign {
						index--
					}
				}
				index++
			}
			if index < 0 {
				index = 0
			}
			if index >= len(grid.grid) {
				index = len(grid.grid) - 1
			}
			newExtrema = append(newExtrema, index)
			if index != currentIndex {
				changes++
			}
		}
		if changes == 0 {
			return nil
		}
		copy(grid.extrema, newExtrema)
	}
	return nil
}

type hilbertInterpolation struct {
	nodes   []float64
	values  []float64
	weights []float64
}

func hilbertRemezDelta(grid *hilbertRemezGrid) (float64, error) {
	extrema := make([]float64, len(grid.extrema))
	for i, index := range grid.extrema {
		extrema[i] = grid.x[index]
	}
	weights := barycentricWeights(extrema)
	numerator := 0.0
	denominator := 0.0
	sign := 1.0
	for i, index := range grid.extrema {
		numerator += weights[i] * grid.desired[index]
		denominator += sign * weights[i] / grid.weight[index]
		sign = -sign
	}
	if denominator == 0 || math.IsNaN(denominator) {
		return 0, fmt.Errorf("Hilbert Remez design produced a singular interpolation system")
	}
	return numerator / denominator, nil
}

func hilbertInterpolatingPolynomial(grid *hilbertRemezGrid, delta float64) (*hilbertInterpolation, error) {
	count := len(grid.extrema) - 1
	nodes := make([]float64, count)
	values := make([]float64, count)
	sign := 1.0
	for i := 0; i < count; i++ {
		index := grid.extrema[i]
		nodes[i] = grid.x[index]
		values[i] = grid.desired[index] - sign*delta/grid.weight[index]
		sign = -sign
	}
	if count == 0 {
		return nil, fmt.Errorf("Hilbert Remez design has no interpolation nodes")
	}
	return &hilbertInterpolation{
		nodes:   nodes,
		values:  values,
		weights: barycentricWeights(nodes),
	}, nil
}

func barycentricWeights(nodes []float64) []float64 {
	weights := make([]float64, len(nodes))
	for i, node := range nodes {
		product := 1.0
		for j, other := range nodes {
			if i != j {
				product *= node - other
			}
		}
		weights[i] = 1 / product
	}
	return weights
}

func evaluateBarycentric(nodes, values, weights []float64, value float64) float64 {
	numerator := 0.0
	denominator := 0.0
	for i, node := range nodes {
		if value == node {
			return values[i]
		}
		term := weights[i] / (value - node)
		numerator += term * values[i]
		denominator += term
	}
	return numerator / denominator
}

func calculateHilbertCosineCoefficients(grid *hilbertRemezGrid, order int) []float32 {
	nfft := 64
	for nfft < 2*order+1 {
		nfft *= 2
	}
	packed := make([]float32, nfft)
	polynomial, _ := hilbertInterpolatingPolynomial(grid, 0)
	// The last Remez approximation is represented by the current extrema.
	// Recompute it with its actual deviation rather than using a zero error.
	if delta, err := hilbertRemezDelta(grid); err == nil {
		polynomial, _ = hilbertInterpolatingPolynomial(grid, delta)
	}
	for i := 0; i <= nfft/2; i++ {
		value := evaluateBarycentric(polynomial.nodes, polynomial.values, polynomial.weights, math.Cos(2*math.Pi*float64(i)/float64(nfft)))
		packed[i] = float32(value)
	}
	sequence := make([]float32, nfft)
	for n := 0; n < nfft; n++ {
		value := float64(packed[0])
		if n%2 == 0 {
			value += float64(packed[nfft/2])
		} else {
			value -= float64(packed[nfft/2])
		}
		for k := 1; k < nfft/2; k++ {
			value += 2 * float64(packed[k]) * math.Cos(2*math.Pi*float64(k*n)/float64(nfft))
		}
		sequence[n] = float32(value / float64(nfft))
	}
	return sequence
}

func interpretHilbertCoefficients(coefficients []float32, order int) []float64 {
	shift := order - 1
	shift %= len(coefficients)
	if shift < 0 {
		shift += len(coefficients)
	}
	shifted := make([]float32, len(coefficients))
	copy(shifted, coefficients[len(coefficients)-shift:])
	copy(shifted[shift:], coefficients[:len(coefficients)-shift])
	coefficientCount := 2*order + 1
	output := make([]float64, coefficientCount)
	output[0] = float64(float32(-0.5 * float64(shifted[0])))
	output[1] = float64(float32(-0.5 * float64(shifted[1])))
	for i := 2; i <= coefficientCount-3; i++ {
		output[i] = float64(float32(0.5 * float64(shifted[i-2]-shifted[i])))
	}
	output[coefficientCount-2] = float64(float32(0.5 * float64(shifted[coefficientCount-4])))
	output[coefficientCount-1] = float64(float32(0.5 * float64(shifted[coefficientCount-3])))
	return output
}

func signOf(value float64) int {
	if value > 0 {
		return 1
	}
	if value < 0 {
		return -1
	}
	return 0
}

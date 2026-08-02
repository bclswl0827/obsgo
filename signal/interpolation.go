package signal

import (
	"fmt"
	"math"
)

// InterpolationMethod names the interpolation kernels exposed by ObsPy.
type InterpolationMethod int

const (
	LinearInterpolation InterpolationMethod = iota
	CubicInterpolation
	NearestInterpolation
	ZeroInterpolation
	SLinearInterpolation
	QuadraticInterpolation
	HermiteInterpolation
	WeightedAverageSlopesInterpolation
	LanczosInterpolation
)

// LanczosKernelWindow selects the window multiplied with the sinc kernel.
type LanczosKernelWindow int

const (
	LanczosWindow LanczosKernelWindow = iota
	HanningLanczosWindow
	BlackmanLanczosWindow
)

// InterpolationOptions controls kernels that need additional parameters.
// LanczosWidth defaults to 20, matching the usual ObsPy recommendation.
// AntiAlias is opt-in: ObsPy interpolation itself does not silently low-pass
// data before downsampling.
type InterpolationOptions struct {
	LanczosWidth   int
	LanczosWindow  LanczosKernelWindow
	AntiAlias      bool
	AntiAliasOrder int
}

// Interpolate evaluates a uniformly sampled sequence at fractional sample
// positions. Positions are measured in input-sample units and must be inside
// the input interval.
func Interpolate(data, positions []float64, method InterpolationMethod) ([]float64, error) {
	return InterpolateWithOptions(data, positions, method, InterpolationOptions{})
}

// InterpolateWithOptions is Interpolate with explicit kernel parameters.
func InterpolateWithOptions(data, positions []float64, method InterpolationMethod, options InterpolationOptions) ([]float64, error) {
	if len(data) == 0 {
		if len(positions) == 0 {
			return []float64{}, nil
		}
		return nil, fmt.Errorf("cannot interpolate an empty sequence")
	}
	if !validInterpolationMethod(method) {
		return nil, fmt.Errorf("unsupported interpolation method %d", method)
	}
	for _, position := range positions {
		if math.IsNaN(position) || math.IsInf(position, 0) || position < 0 || position > float64(len(data)-1) {
			return nil, fmt.Errorf("interpolation position %.6g is outside [0, %d]", position, len(data)-1)
		}
	}
	return interpolatePositions(data, positions, method, options)
}

// Interpolate1D follows ObsPy's interpolate_1d parameter model. Times are
// expressed in the same arbitrary unit; normally they are seconds.
func Interpolate1D(data []float64, oldStart, oldDelta, newStart, newDelta float64, newNpts int, method InterpolationMethod) ([]float64, error) {
	return Interpolate1DWithOptions(data, oldStart, oldDelta, newStart, newDelta, newNpts, method, InterpolationOptions{})
}

// Interpolate1DWithOptions adds Lanczos and optional anti-aliasing controls to
// Interpolate1D. It never extrapolates beyond the old data interval.
func Interpolate1DWithOptions(data []float64, oldStart, oldDelta, newStart, newDelta float64, newNpts int, method InterpolationMethod, options InterpolationOptions) ([]float64, error) {
	if _, _, err := validateInterpolationGrid(data, oldStart, oldDelta, newStart, newDelta, newNpts, method); err != nil {
		return nil, err
	}
	working := append([]float64(nil), data...)
	if options.AntiAlias && newDelta > oldDelta {
		order := options.AntiAliasOrder
		if order <= 0 {
			order = 4
		}
		filtered, filterErr := Lowpass(working, 0.45/newDelta, 1/oldDelta, order, true)
		if filterErr != nil {
			return nil, fmt.Errorf("anti-alias lowpass: %w", filterErr)
		}
		working = filtered
	}

	positions := make([]float64, newNpts)
	position := (newStart - oldStart) / oldDelta
	step := newDelta / oldDelta
	for index := range positions {
		positions[index] = position + float64(index)*step
		if math.Abs(positions[index]) < 1e-12 {
			positions[index] = 0
		}
		if math.Abs(positions[index]-float64(len(data)-1)) < 1e-12 {
			positions[index] = float64(len(data) - 1)
		}
	}
	output, err := interpolatePositions(working, positions, method, options)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Resample is a convenience wrapper preserving the input's first and last
// sample times. It does not apply anti-aliasing implicitly.
func Resample(data []float64, inputRate, outputRate float64, method InterpolationMethod) ([]float64, error) {
	if !finitePositive(inputRate) || !finitePositive(outputRate) {
		return nil, fmt.Errorf("sample rates must be finite and positive")
	}
	if len(data) == 0 {
		return []float64{}, nil
	}
	if len(data) == 1 {
		return []float64{data[0]}, nil
	}
	count := int(math.Round(float64(len(data)-1)*outputRate/inputRate)) + 1
	if count < 1 {
		count = 1
	}
	return Interpolate1D(data, 0, 1/inputRate, 0, 1/outputRate, count, method)
}

// ResampleTo is the explicit resampling API. It allows the caller to choose
// the time origin, output point count and anti-aliasing policy.
func ResampleTo(data []float64, oldStart, inputRate, newStart, outputRate float64, newNpts int, method InterpolationMethod, options InterpolationOptions) ([]float64, error) {
	if !finitePositive(inputRate) || !finitePositive(outputRate) {
		return nil, fmt.Errorf("sample rates must be finite and positive")
	}
	return Interpolate1DWithOptions(data, oldStart, 1/inputRate, newStart, 1/outputRate, newNpts, method, options)
}

func validateInterpolationGrid(data []float64, oldStart, oldDelta, newStart, newDelta float64, newNpts int, method InterpolationMethod) (float64, float64, error) {
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("cannot interpolate an empty sequence")
	}
	if !finitePositive(oldDelta) || !finitePositive(newDelta) {
		return 0, 0, fmt.Errorf("time steps must be finite and positive")
	}
	if math.IsNaN(oldStart) || math.IsInf(oldStart, 0) || math.IsNaN(newStart) || math.IsInf(newStart, 0) {
		return 0, 0, fmt.Errorf("start times must be finite")
	}
	if newNpts <= 0 {
		return 0, 0, fmt.Errorf("number of output points must be positive")
	}
	if !validInterpolationMethod(method) {
		return 0, 0, fmt.Errorf("unsupported interpolation method %d", method)
	}
	oldEnd := oldStart + oldDelta*float64(len(data)-1)
	newEnd := newStart + newDelta*float64(newNpts-1)
	if newStart < oldStart-1e-12 || newEnd > oldEnd+1e-12 {
		return 0, 0, fmt.Errorf("new interpolation interval [%g, %g] is outside old interval [%g, %g]", newStart, newEnd, oldStart, oldEnd)
	}
	return oldEnd, newEnd, nil
}

func interpolatePositions(data, positions []float64, method InterpolationMethod, options InterpolationOptions) ([]float64, error) {
	var spline *splineInterpolator
	if method == QuadraticInterpolation || method == CubicInterpolation {
		var err error
		degree := 2
		if method == CubicInterpolation {
			degree = 3
		}
		spline, err = newSplineInterpolator(data, degree)
		if err != nil {
			return nil, err
		}
	}
	var slopes []float64
	switch method {
	case HermiteInterpolation:
		slopes = centralSlopes(data)
	case WeightedAverageSlopesInterpolation:
		slopes = weightedAverageSlopes(data)
	}
	if method == LanczosInterpolation {
		if options.LanczosWidth <= 0 {
			options.LanczosWidth = 20
		}
		if options.LanczosWindow < LanczosWindow || options.LanczosWindow > BlackmanLanczosWindow {
			return nil, fmt.Errorf("unsupported Lanczos window %d", options.LanczosWindow)
		}
	}
	output := make([]float64, len(positions))
	for index, position := range positions {
		switch method {
		case LinearInterpolation, SLinearInterpolation:
			output[index] = interpolateLinear(data, position)
		case NearestInterpolation:
			output[index] = data[nearestIndex(position, len(data))]
		case ZeroInterpolation:
			output[index] = data[int(math.Floor(position))]
		case QuadraticInterpolation, CubicInterpolation:
			output[index] = spline.evaluate(position)
		case HermiteInterpolation, WeightedAverageSlopesInterpolation:
			output[index] = interpolateHermite(data, slopes, position)
		case LanczosInterpolation:
			output[index] = interpolateLanczos(data, position, options.LanczosWidth, options.LanczosWindow)
		}
	}
	return output, nil
}

func validInterpolationMethod(method InterpolationMethod) bool {
	return method >= LinearInterpolation && method <= LanczosInterpolation
}

func finitePositive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func interpolateLinear(data []float64, position float64) float64 {
	left := int(math.Floor(position))
	if left >= len(data)-1 {
		return data[len(data)-1]
	}
	fraction := position - float64(left)
	return data[left] + fraction*(data[left+1]-data[left])
}

func nearestIndex(position float64, length int) int {
	index := int(math.Ceil(position - 0.5))
	if index < 0 {
		return 0
	}
	if index >= length {
		return length - 1
	}
	return index
}

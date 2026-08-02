package signal

import (
	"fmt"
	"math"
)

// splineInterpolator is the not-a-knot interpolating B-spline used by
// scipy.interpolate.InterpolatedUnivariateSpline for uniformly sampled data.
type splineInterpolator struct {
	degree       int
	knots        []float64
	coefficients []float64
	last         float64
}

func newSplineInterpolator(data []float64, degree int) (*splineInterpolator, error) {
	if degree < 1 || degree > 3 {
		return nil, fmt.Errorf("spline degree must be between one and three")
	}
	if len(data) < degree+1 {
		return nil, fmt.Errorf("degree-%d spline interpolation requires at least %d samples", degree, degree+1)
	}
	knots := notAKnotVector(len(data), degree)
	matrix := make([][]float64, len(data))
	for row := range matrix {
		matrix[row] = make([]float64, len(data))
		for column := range matrix[row] {
			matrix[row][column] = bsplineBasis(column, degree, float64(row), knots, len(data))
		}
	}
	coefficients, err := solveDenseInterpolationSystem(matrix, append([]float64(nil), data...))
	if err != nil {
		return nil, fmt.Errorf("degree-%d spline: %w", degree, err)
	}
	return &splineInterpolator{
		degree:       degree,
		knots:        knots,
		coefficients: coefficients,
		last:         float64(len(data) - 1),
	}, nil
}

func notAKnotVector(length, degree int) []float64 {
	last := float64(length - 1)
	knots := make([]float64, 0, length+degree+1)
	for index := 0; index <= degree; index++ {
		knots = append(knots, 0)
	}
	if degree%2 == 1 {
		for index := (degree + 1) / 2; index < length-(degree+1)/2; index++ {
			knots = append(knots, float64(index))
		}
	} else {
		for index := degree / 2; index < length-degree/2-1; index++ {
			knots = append(knots, float64(index)+0.5)
		}
	}
	for index := 0; index <= degree; index++ {
		knots = append(knots, last)
	}
	return knots
}

func bsplineBasis(index, degree int, value float64, knots []float64, basisCount int) float64 {
	if degree == 0 {
		if knots[index] <= value && value < knots[index+1] || value == knots[len(knots)-1] && index == basisCount-1 {
			return 1
		}
		return 0
	}
	left := 0.0
	leftDenominator := knots[index+degree] - knots[index]
	if leftDenominator != 0 {
		left = (value - knots[index]) / leftDenominator * bsplineBasis(index, degree-1, value, knots, basisCount)
	}
	right := 0.0
	rightDenominator := knots[index+degree+1] - knots[index+1]
	if rightDenominator != 0 {
		right = (knots[index+degree+1] - value) / rightDenominator * bsplineBasis(index+1, degree-1, value, knots, basisCount)
	}
	return left + right
}

func (spline *splineInterpolator) evaluate(position float64) float64 {
	if position == spline.last {
		return spline.coefficients[len(spline.coefficients)-1]
	}
	value := 0.0
	for index, coefficient := range spline.coefficients {
		value += coefficient * bsplineBasis(index, spline.degree, position, spline.knots, len(spline.coefficients))
	}
	return value
}

func solveDenseInterpolationSystem(matrix [][]float64, right []float64) ([]float64, error) {
	n := len(right)
	for pivot := 0; pivot < n; pivot++ {
		best := pivot
		for row := pivot + 1; row < n; row++ {
			if math.Abs(matrix[row][pivot]) > math.Abs(matrix[best][pivot]) {
				best = row
			}
		}
		if math.Abs(matrix[best][pivot]) <= 1e-14 {
			return nil, fmt.Errorf("singular interpolation system")
		}
		matrix[pivot], matrix[best] = matrix[best], matrix[pivot]
		right[pivot], right[best] = right[best], right[pivot]
		for row := pivot + 1; row < n; row++ {
			factor := matrix[row][pivot] / matrix[pivot][pivot]
			for column := pivot; column < n; column++ {
				matrix[row][column] -= factor * matrix[pivot][column]
			}
			right[row] -= factor * right[pivot]
		}
	}
	result := make([]float64, n)
	for row := n - 1; row >= 0; row-- {
		value := right[row]
		for column := row + 1; column < n; column++ {
			value -= matrix[row][column] * result[column]
		}
		result[row] = value / matrix[row][row]
	}
	return result, nil
}

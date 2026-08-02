package signal

import (
	"fmt"
	"math"
	"sort"
)

// PolarizationResult describes the Flinn eigenstructure of a three-component
// window. Components in Principal are ordered Z, North, East. Azimuth follows
// ObsPy's convention: clockwise from North and folded into [0, 180] degrees;
// incidence is measured from the vertical and folded into [0, 90] degrees.
type PolarizationResult struct {
	Eigenvalues    [3]float64
	Principal      [3]float64 // Z, North, East components of the principal vector.
	Azimuth        float64
	Incidence      float64
	AzimuthError   float64
	IncidenceError float64
	Rectilinearity float64
	Planarity      float64
	Ellipticity    float64
}

// AnalyzePolarization estimates the covariance-based polarization of one
// three-component window using the principal eigenvector.
func AnalyzePolarization(vertical, north, east []float64) (PolarizationResult, error) {
	if len(vertical) != len(north) || len(vertical) != len(east) {
		return PolarizationResult{}, fmt.Errorf("Z, north, and east components must have the same length")
	}
	if len(vertical) < 2 {
		return PolarizationResult{}, fmt.Errorf("polarization requires at least two samples")
	}
	meanZ, meanN, meanE := Mean(vertical), Mean(north), Mean(east)
	var covariance [3][3]float64
	for index := range vertical {
		value := [3]float64{vertical[index] - meanZ, north[index] - meanN, east[index] - meanE}
		for row := 0; row < 3; row++ {
			for column := row; column < 3; column++ {
				covariance[row][column] += value[row] * value[column]
			}
		}
	}
	for row := 0; row < 3; row++ {
		for column := row; column < 3; column++ {
			covariance[row][column] /= float64(len(vertical) - 1)
			covariance[column][row] = covariance[row][column]
		}
	}
	eigenvalues, eigenvectors := symmetricEigen3(covariance)
	result := PolarizationResult{Eigenvalues: eigenvalues}
	for component := 0; component < 3; component++ {
		result.Principal[component] = eigenvectors[component][0]
	}
	result.Azimuth = math.Atan2(result.Principal[2], result.Principal[1]) / DegreesToRadians
	if result.Azimuth < 0 {
		result.Azimuth += 360
	}
	result.Incidence = math.Atan2(math.Hypot(result.Principal[1], result.Principal[2]), result.Principal[0]) / DegreesToRadians
	if result.Incidence < 0 {
		result.Incidence += 180
	}
	if result.Incidence > 90 {
		result.Incidence = 180 - result.Incidence
		if result.Azimuth > 180 {
			result.Azimuth -= 180
		} else {
			result.Azimuth += 180
		}
	}
	if result.Azimuth > 180 {
		result.Azimuth -= 180
	}
	if eigenvalues[0] > 0 {
		result.Rectilinearity = clamp01(1 - math.Sqrt(eigenvalues[1]/eigenvalues[0]))
		result.Ellipticity = math.Sqrt(eigenvalues[1] / eigenvalues[0])
	}
	if eigenvalues[0]+eigenvalues[1] > 0 {
		result.Planarity = clamp01(1 - 2*eigenvalues[2]/(eigenvalues[0]+eigenvalues[1]))
	}
	return result, nil
}

// PolarizationAnalysis applies AnalyzePolarization to complete, equally
// sized windows. The final incomplete window is omitted.
func PolarizationAnalysis(vertical, north, east []float64, windowLength, step int) ([]PolarizationResult, error) {
	if windowLength < 2 {
		return nil, fmt.Errorf("polarization window must contain at least two samples")
	}
	if step <= 0 {
		step = windowLength
	}
	if len(vertical) != len(north) || len(vertical) != len(east) {
		return nil, fmt.Errorf("Z, north, and east components must have the same length")
	}
	results := make([]PolarizationResult, 0)
	for start := 0; start+windowLength <= len(vertical); start += step {
		result, err := AnalyzePolarization(vertical[start:start+windowLength], north[start:start+windowLength], east[start:start+windowLength])
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func symmetricEigen3(matrix [3][3]float64) ([3]float64, [3][3]float64) {
	values := matrix
	vectors := [3][3]float64{}
	for index := 0; index < 3; index++ {
		vectors[index][index] = 1
	}
	for iteration := 0; iteration < 50; iteration++ {
		p, q := 0, 1
		if math.Abs(values[0][2]) > math.Abs(values[p][q]) {
			p, q = 0, 2
		}
		if math.Abs(values[1][2]) > math.Abs(values[p][q]) {
			p, q = 1, 2
		}
		if math.Abs(values[p][q]) < 1e-15 {
			break
		}
		tau := (values[q][q] - values[p][p]) / (2 * values[p][q])
		t := math.Copysign(1/(math.Abs(tau)+math.Sqrt(1+tau*tau)), tau)
		c := 1 / math.Sqrt(1+t*t)
		s := t * c
		app, aqq, apq := values[p][p], values[q][q], values[p][q]
		values[p][p] = c*c*app - 2*s*c*apq + s*s*aqq
		values[q][q] = s*s*app + 2*s*c*apq + c*c*aqq
		values[p][q], values[q][p] = 0, 0
		for index := 0; index < 3; index++ {
			if index == p || index == q {
				continue
			}
			vip, viq := values[index][p], values[index][q]
			values[index][p], values[p][index] = c*vip-s*viq, c*vip-s*viq
			values[index][q], values[q][index] = s*vip+c*viq, s*vip+c*viq
		}
		for index := 0; index < 3; index++ {
			vip, viq := vectors[index][p], vectors[index][q]
			vectors[index][p], vectors[index][q] = c*vip-s*viq, s*vip+c*viq
		}
	}
	indices := []int{0, 1, 2}
	sort.Slice(indices, func(left, right int) bool {
		return values[indices[left]][indices[left]] > values[indices[right]][indices[right]]
	})
	var eigenvalues [3]float64
	var eigenvectors [3][3]float64
	for output, source := range indices {
		eigenvalues[output] = math.Max(0, values[source][source])
		for component := 0; component < 3; component++ {
			eigenvectors[component][output] = vectors[component][source]
		}
	}
	return eigenvalues, eigenvectors
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

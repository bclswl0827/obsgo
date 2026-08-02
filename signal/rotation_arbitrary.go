package signal

import (
	"fmt"
	"math"
)

// ComponentOrientation describes a SEED component orientation. Azimuth is
// measured clockwise from North and dip is measured down from horizontal.
type ComponentOrientation struct {
	Azimuth float64
	Dip     float64
}

// Rotate2ZNE converts three arbitrarily oriented components to ZNE. When
// inverse is true, the input components are Z, N, E and the output components
// use the three supplied orientations. The orientations need only be linearly
// independent; they do not need to be orthogonal.
func Rotate2ZNE(data1 []float64, azimuth1, dip1 float64, data2 []float64, azimuth2, dip2 float64, data3 []float64, azimuth3, dip3 float64, inverse bool) (z, north, east []float64, err error) {
	if len(data1) != len(data2) || len(data1) != len(data3) {
		return nil, nil, nil, fmt.Errorf("all three data arrays must have the same length")
	}
	matrix := [3][3]float64{
		dipAzimuthBaseVector(azimuth1, dip1),
		dipAzimuthBaseVector(azimuth2, dip2),
		dipAzimuthBaseVector(azimuth3, dip3),
	}
	determinant := determinant3(matrix)
	if math.IsNaN(determinant) || math.IsInf(determinant, 0) || math.Abs(determinant) <= 1e-6 || math.Abs(determinant) >= 1e6 {
		return nil, nil, nil, fmt.Errorf("component orientations are not linearly independent; determinant is %g", determinant)
	}
	if !inverse {
		var inverseMatrix [3][3]float64
		inverseMatrix, err = inverse3(matrix)
		if err != nil {
			return nil, nil, nil, err
		}
		matrix = inverseMatrix
	}
	inputs := [3][]float64{data1, data2, data3}
	outputs := [3][]float64{make([]float64, len(data1)), make([]float64, len(data1)), make([]float64, len(data1))}
	for sample := range data1 {
		for output := 0; output < 3; output++ {
			value := matrix[output][0]*inputs[0][sample] + matrix[output][1]*inputs[1][sample] + matrix[output][2]*inputs[2][sample]
			if value == 0 {
				value = 0
			}
			outputs[output][sample] = value
		}
	}
	return outputs[0], outputs[1], outputs[2], nil
}

// RotateArbitraryToZNE is the structured form of Rotate2ZNE for arbitrary
// component arrays and orientations.
func RotateArbitraryToZNE(data [3][]float64, orientations [3]ComponentOrientation) (z, north, east []float64, err error) {
	return Rotate2ZNE(
		data[0], orientations[0].Azimuth, orientations[0].Dip,
		data[1], orientations[1].Azimuth, orientations[1].Dip,
		data[2], orientations[2].Azimuth, orientations[2].Dip,
		false,
	)
}

// RotateZNEToArbitrary converts ZNE data into the three supplied component
// orientations.
func RotateZNEToArbitrary(z, north, east []float64, orientations [3]ComponentOrientation) (data [3][]float64, err error) {
	data[0], data[1], data[2], err = Rotate2ZNE(
		z, orientations[0].Azimuth, orientations[0].Dip,
		north, orientations[1].Azimuth, orientations[1].Dip,
		east, orientations[2].Azimuth, orientations[2].Dip,
		true,
	)
	return data, err
}

func dipAzimuthBaseVector(azimuth, dip float64) [3]float64 {
	dipRadians := dip * DegreesToRadians
	azimuthRadians := azimuth * DegreesToRadians
	return [3]float64{
		-math.Sin(dipRadians),
		math.Cos(azimuthRadians) * math.Cos(dipRadians),
		math.Sin(azimuthRadians) * math.Cos(dipRadians),
	}
}

func determinant3(matrix [3][3]float64) float64 {
	return matrix[0][0]*(matrix[1][1]*matrix[2][2]-matrix[1][2]*matrix[2][1]) -
		matrix[0][1]*(matrix[1][0]*matrix[2][2]-matrix[1][2]*matrix[2][0]) +
		matrix[0][2]*(matrix[1][0]*matrix[2][1]-matrix[1][1]*matrix[2][0])
}

func inverse3(matrix [3][3]float64) ([3][3]float64, error) {
	determinant := determinant3(matrix)
	if math.Abs(determinant) <= 1e-15 {
		return [3][3]float64{}, fmt.Errorf("component orientation matrix is singular")
	}
	return [3][3]float64{
		{
			(matrix[1][1]*matrix[2][2] - matrix[1][2]*matrix[2][1]) / determinant,
			(matrix[0][2]*matrix[2][1] - matrix[0][1]*matrix[2][2]) / determinant,
			(matrix[0][1]*matrix[1][2] - matrix[0][2]*matrix[1][1]) / determinant,
		},
		{
			(matrix[1][2]*matrix[2][0] - matrix[1][0]*matrix[2][2]) / determinant,
			(matrix[0][0]*matrix[2][2] - matrix[0][2]*matrix[2][0]) / determinant,
			(matrix[0][2]*matrix[1][0] - matrix[0][0]*matrix[1][2]) / determinant,
		},
		{
			(matrix[1][0]*matrix[2][1] - matrix[1][1]*matrix[2][0]) / determinant,
			(matrix[0][1]*matrix[2][0] - matrix[0][0]*matrix[2][1]) / determinant,
			(matrix[0][0]*matrix[1][1] - matrix[0][1]*matrix[1][0]) / determinant,
		},
	}, nil
}

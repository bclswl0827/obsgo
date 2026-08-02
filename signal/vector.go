package signal

import (
	"fmt"
	"math"
)

const DegreesToRadians = math.Pi / 180

func Rotate(componentA []float64, azimuthA float64, componentB []float64, azimuthB, outputAzimuth float64) (radial, transverse []float64, err error) {
	if len(componentA) != len(componentB) {
		return nil, nil, fmt.Errorf("components must have the same length")
	}
	azimuthDifference := math.Mod(azimuthB-azimuthA+360, 360)
	if math.Abs(azimuthDifference-90) > 1e-9 {
		return nil, nil, fmt.Errorf("component B azimuth must be component A azimuth plus 90 degrees")
	}
	rotation := (outputAzimuth - azimuthA) * DegreesToRadians
	cosine := math.Cos(rotation)
	sine := math.Sin(rotation)
	radial = make([]float64, len(componentA))
	transverse = make([]float64, len(componentA))
	for index := range componentA {
		transverse[index] = cosine*componentB[index] - sine*componentA[index]
		radial[index] = sine*componentB[index] + cosine*componentA[index]
	}
	return radial, transverse, nil
}

func VectorMagnitude(componentA, componentB, componentC []float64) ([]float64, error) {
	if len(componentA) != len(componentB) || len(componentA) != len(componentC) {
		return nil, fmt.Errorf("components must have the same length")
	}
	output := make([]float64, len(componentA))
	for index := range output {
		output[index] = math.Sqrt(componentA[index]*componentA[index] + componentB[index]*componentB[index] + componentC[index]*componentC[index])
	}
	return output, nil
}

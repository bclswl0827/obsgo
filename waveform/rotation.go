package waveform

import (
	"fmt"

	"github.com/bclswl0827/obsgo/signal"
)

// RotateZNEToZRT rotates aligned three-component waveform segments to ZRT.
func RotateZNEToZRT(vertical, north, east Seismogram, backAzimuth float64) (z, radial, transverse Seismogram, err error) {
	return rotateThree(vertical, north, east,
		func(a, b, c []float64) ([]float64, []float64, []float64, error) {
			return signal.RotateZNEToZRT(a, b, c, backAzimuth)
		}, vertical.Unit, vertical.SourceID, north.SourceID, east.SourceID)
}

// RotateZRTToZNE rotates aligned ZRT waveform segments back to ZNE.
func RotateZRTToZNE(vertical, radial, transverse Seismogram, backAzimuth float64) (z, north, east Seismogram, err error) {
	return rotateThree(vertical, radial, transverse,
		func(a, b, c []float64) ([]float64, []float64, []float64, error) {
			return signal.RotateZRTToZNE(a, b, c, backAzimuth)
		}, vertical.Unit, vertical.SourceID, radial.SourceID, transverse.SourceID)
}

// RotateZNEToLQT rotates aligned three-component waveform segments to LQT.
func RotateZNEToLQT(vertical, north, east Seismogram, backAzimuth, inclination float64) (longitudinal, transverseQ, transverseT Seismogram, err error) {
	return rotateThree(vertical, north, east,
		func(a, b, c []float64) ([]float64, []float64, []float64, error) {
			return signal.RotateZNEToLQT(a, b, c, backAzimuth, inclination)
		}, vertical.Unit, vertical.SourceID, north.SourceID, east.SourceID)
}

// RotateLQTToZNE rotates aligned LQT waveform segments back to ZNE.
func RotateLQTToZNE(longitudinal, transverseQ, transverseT Seismogram, backAzimuth, inclination float64) (vertical, north, east Seismogram, err error) {
	return rotateThree(longitudinal, transverseQ, transverseT,
		func(a, b, c []float64) ([]float64, []float64, []float64, error) {
			return signal.RotateLQTToZNE(a, b, c, backAzimuth, inclination)
		}, longitudinal.Unit, longitudinal.SourceID, transverseQ.SourceID, transverseT.SourceID)
}

// RotateArbitraryToZNE applies the full non-orthogonal SEED orientation
// matrix to three waveform components.
func RotateArbitraryToZNE(first, second, third Seismogram, orientations [3]signal.ComponentOrientation) (z, north, east Seismogram, err error) {
	return rotateThree(first, second, third,
		func(a, b, c []float64) ([]float64, []float64, []float64, error) {
			return signal.RotateArbitraryToZNE([3][]float64{a, b, c}, orientations)
		}, first.Unit, first.SourceID, second.SourceID, third.SourceID)
}

// RotateZNEToArbitrary converts ZNE waveform segments to the requested three
// arbitrary component orientations.
func RotateZNEToArbitrary(vertical, north, east Seismogram, orientations [3]signal.ComponentOrientation) (first, second, third Seismogram, err error) {
	return rotateThree(vertical, north, east,
		func(a, b, c []float64) ([]float64, []float64, []float64, error) {
			values, rotateErr := signal.RotateZNEToArbitrary(a, b, c, orientations)
			return values[0], values[1], values[2], rotateErr
		}, vertical.Unit, vertical.SourceID, north.SourceID, east.SourceID)
}

type rotateThreeFunc func([]float64, []float64, []float64) ([]float64, []float64, []float64, error)

func rotateThree(first, second, third Seismogram, operation rotateThreeFunc, unit, firstSource, secondSource, thirdSource string) (Seismogram, Seismogram, Seismogram, error) {
	if operation == nil {
		return Seismogram{}, Seismogram{}, Seismogram{}, fmt.Errorf("waveform rotation operation is nil")
	}
	segments, err := alignTriples(first, second, third)
	if err != nil {
		return Seismogram{}, Seismogram{}, Seismogram{}, err
	}
	outputs := [3][]SeismogramSegment{}
	for _, segment := range segments {
		rotatedFirst, rotatedSecond, rotatedThird, err := operation(segment.Data[0], segment.Data[1], segment.Data[2])
		if err != nil {
			return Seismogram{}, Seismogram{}, Seismogram{}, err
		}
		values := [3][]float64{rotatedFirst, rotatedSecond, rotatedThird}
		sourceIDs := [3]string{firstSource, secondSource, thirdSource}
		for component := range outputs {
			outputs[component] = append(outputs[component], SeismogramSegment{
				Data:       values[component],
				StartTime:  segment.StartTime,
				SampleRate: segment.SampleRate,
				Unit:       unit,
				SourceID:   sourceIDs[component],
			})
		}
	}
	firstResult, err := buildSeismogram(outputs[0], unit, firstSource)
	if err != nil {
		return Seismogram{}, Seismogram{}, Seismogram{}, err
	}
	secondResult, err := buildSeismogram(outputs[1], unit, secondSource)
	if err != nil {
		return Seismogram{}, Seismogram{}, Seismogram{}, err
	}
	thirdResult, err := buildSeismogram(outputs[2], unit, thirdSource)
	if err != nil {
		return Seismogram{}, Seismogram{}, Seismogram{}, err
	}
	return firstResult, secondResult, thirdResult, nil
}

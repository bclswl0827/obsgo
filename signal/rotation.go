package signal

import (
	"fmt"
	"math"
)

// RotateNEToRT converts North/East components to radial/transverse
// components using the ObsPy back-azimuth convention.
func RotateNEToRT(north, east []float64, backAzimuth float64) (radial, transverse []float64, err error) {
	if len(north) != len(east) {
		return nil, nil, fmt.Errorf("north and east components must have the same length")
	}
	if backAzimuth < 0 || backAzimuth > 360 {
		return nil, nil, fmt.Errorf("back azimuth must be between 0 and 360 degrees")
	}
	angle := backAzimuth * DegreesToRadians
	cosine, sine := math.Cos(angle), math.Sin(angle)
	radial = make([]float64, len(north))
	transverse = make([]float64, len(north))
	for index := range north {
		radial[index] = -north[index]*cosine - east[index]*sine
		transverse[index] = north[index]*sine - east[index]*cosine
	}
	return radial, transverse, nil
}

// RotateRTToNE is the inverse of RotateNEToRT.
func RotateRTToNE(radial, transverse []float64, backAzimuth float64) (north, east []float64, err error) {
	if len(radial) != len(transverse) {
		return nil, nil, fmt.Errorf("radial and transverse components must have the same length")
	}
	if backAzimuth < 0 || backAzimuth > 360 {
		return nil, nil, fmt.Errorf("back azimuth must be between 0 and 360 degrees")
	}
	angle := backAzimuth * DegreesToRadians
	cosine, sine := math.Cos(angle), math.Sin(angle)
	north = make([]float64, len(radial))
	east = make([]float64, len(radial))
	for index := range radial {
		north[index] = -radial[index]*cosine + transverse[index]*sine
		east[index] = -radial[index]*sine - transverse[index]*cosine
	}
	return north, east, nil
}

// RotateZNEToZRT rotates the horizontal components and preserves Z.
func RotateZNEToZRT(vertical, north, east []float64, backAzimuth float64) (z, radial, transverse []float64, err error) {
	if len(vertical) != len(north) || len(vertical) != len(east) {
		return nil, nil, nil, fmt.Errorf("Z, north, and east components must have the same length")
	}
	radial, transverse, err = RotateNEToRT(north, east, backAzimuth)
	if err != nil {
		return nil, nil, nil, err
	}
	return append([]float64(nil), vertical...), radial, transverse, nil
}

// RotateZRTToZNE is the inverse of RotateZNEToZRT.
func RotateZRTToZNE(vertical, radial, transverse []float64, backAzimuth float64) (z, north, east []float64, err error) {
	if len(vertical) != len(radial) || len(vertical) != len(transverse) {
		return nil, nil, nil, fmt.Errorf("Z, radial, and transverse components must have the same length")
	}
	north, east, err = RotateRTToNE(radial, transverse, backAzimuth)
	if err != nil {
		return nil, nil, nil, err
	}
	return append([]float64(nil), vertical...), north, east, nil
}

// RotateZNEToLQT converts ZNE components to ray coordinates using the ObsPy
// convention: inclination is measured from the upward vertical and may span
// 0..360 degrees.
func RotateZNEToLQT(vertical, north, east []float64, backAzimuth, inclination float64) (longitudinal, transverseQ, transverseT []float64, err error) {
	if len(vertical) != len(north) || len(vertical) != len(east) {
		return nil, nil, nil, fmt.Errorf("Z, north, and east components must have the same length")
	}
	if backAzimuth < 0 || backAzimuth > 360 || inclination < 0 || inclination > 360 {
		return nil, nil, nil, fmt.Errorf("back azimuth and inclination must be between 0 and 360 degrees")
	}
	ba := backAzimuth * DegreesToRadians
	inc := inclination * DegreesToRadians
	cb, sb := math.Cos(ba), math.Sin(ba)
	ci, si := math.Cos(inc), math.Sin(inc)
	longitudinal = make([]float64, len(vertical))
	transverseQ = make([]float64, len(vertical))
	transverseT = make([]float64, len(vertical))
	for index := range vertical {
		longitudinal[index] = vertical[index]*ci - north[index]*si*cb - east[index]*si*sb
		transverseQ[index] = vertical[index]*si + north[index]*ci*cb + east[index]*ci*sb
		transverseT[index] = north[index]*sb - east[index]*cb
	}
	return longitudinal, transverseQ, transverseT, nil
}

// RotateLQTToZNE is the inverse of RotateZNEToLQT.
func RotateLQTToZNE(longitudinal, transverseQ, transverseT []float64, backAzimuth, inclination float64) (vertical, north, east []float64, err error) {
	if len(longitudinal) != len(transverseQ) || len(longitudinal) != len(transverseT) {
		return nil, nil, nil, fmt.Errorf("L, Q, and T components must have the same length")
	}
	if backAzimuth < 0 || backAzimuth > 360 || inclination < 0 || inclination > 360 {
		return nil, nil, nil, fmt.Errorf("back azimuth and inclination must be between 0 and 360 degrees")
	}
	ba := backAzimuth * DegreesToRadians
	inc := inclination * DegreesToRadians
	cb, sb := math.Cos(ba), math.Sin(ba)
	ci, si := math.Cos(inc), math.Sin(inc)
	vertical = make([]float64, len(longitudinal))
	north = make([]float64, len(longitudinal))
	east = make([]float64, len(longitudinal))
	for index := range longitudinal {
		vertical[index] = longitudinal[index]*ci + transverseQ[index]*si
		north[index] = -longitudinal[index]*si*cb + transverseQ[index]*ci*cb + transverseT[index]*sb
		east[index] = -longitudinal[index]*si*sb + transverseQ[index]*ci*sb - transverseT[index]*cb
	}
	return vertical, north, east, nil
}

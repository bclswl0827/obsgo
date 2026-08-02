package waveform

import "github.com/bclswl0827/obsgo/signal"

func RotateTraceZNEToZRT(vertical, north, east Trace, backAzimuth float64) (z, radial, transverse Trace, err error) {
	resultZ, resultR, resultT, err := RotateZNEToZRT(vertical.Seismogram, north.Seismogram, east.Seismogram, backAzimuth)
	if err != nil {
		return Trace{}, Trace{}, Trace{}, err
	}
	vertical.Seismogram, north.Seismogram, east.Seismogram = resultZ, resultR, resultT
	return withTraceComponent(vertical, 'Z'), withTraceComponent(north, 'R'), withTraceComponent(east, 'T'), nil
}

func RotateTraceZRTToZNE(vertical, radial, transverse Trace, backAzimuth float64) (z, north, east Trace, err error) {
	resultZ, resultN, resultE, err := RotateZRTToZNE(vertical.Seismogram, radial.Seismogram, transverse.Seismogram, backAzimuth)
	if err != nil {
		return Trace{}, Trace{}, Trace{}, err
	}
	vertical.Seismogram, radial.Seismogram, transverse.Seismogram = resultZ, resultN, resultE
	return withTraceComponent(vertical, 'Z'), withTraceComponent(radial, 'N'), withTraceComponent(transverse, 'E'), nil
}

func RotateTraceZNEToLQT(vertical, north, east Trace, backAzimuth, inclination float64) (longitudinal, transverseQ, transverseT Trace, err error) {
	resultL, resultQ, resultT, err := RotateZNEToLQT(vertical.Seismogram, north.Seismogram, east.Seismogram, backAzimuth, inclination)
	if err != nil {
		return Trace{}, Trace{}, Trace{}, err
	}
	vertical.Seismogram, north.Seismogram, east.Seismogram = resultL, resultQ, resultT
	return withTraceComponent(vertical, 'L'), withTraceComponent(north, 'Q'), withTraceComponent(east, 'T'), nil
}

func RotateTraceLQTToZNE(longitudinal, transverseQ, transverseT Trace, backAzimuth, inclination float64) (vertical, north, east Trace, err error) {
	resultZ, resultN, resultE, err := RotateLQTToZNE(longitudinal.Seismogram, transverseQ.Seismogram, transverseT.Seismogram, backAzimuth, inclination)
	if err != nil {
		return Trace{}, Trace{}, Trace{}, err
	}
	longitudinal.Seismogram, transverseQ.Seismogram, transverseT.Seismogram = resultZ, resultN, resultE
	return withTraceComponent(longitudinal, 'Z'), withTraceComponent(transverseQ, 'N'), withTraceComponent(transverseT, 'E'), nil
}

func RotateTraceArbitraryToZNE(first, second, third Trace, orientations [3]signal.ComponentOrientation) (z, north, east Trace, err error) {
	resultZ, resultN, resultE, err := RotateArbitraryToZNE(first.Seismogram, second.Seismogram, third.Seismogram, orientations)
	if err != nil {
		return Trace{}, Trace{}, Trace{}, err
	}
	first.Seismogram, second.Seismogram, third.Seismogram = resultZ, resultN, resultE
	return withTraceComponent(first, 'Z'), withTraceComponent(second, 'N'), withTraceComponent(third, 'E'), nil
}

// RotateTraceZNEToArbitrary retains the three destination Trace identifiers:
// arbitrary orientations do not have a universally inferable SEED component
// code.
func RotateTraceZNEToArbitrary(vertical, north, east Trace, orientations [3]signal.ComponentOrientation) (first, second, third Trace, err error) {
	resultFirst, resultSecond, resultThird, err := RotateZNEToArbitrary(vertical.Seismogram, north.Seismogram, east.Seismogram, orientations)
	if err != nil {
		return Trace{}, Trace{}, Trace{}, err
	}
	vertical.Seismogram, north.Seismogram, east.Seismogram = resultFirst, resultSecond, resultThird
	return vertical, north, east, nil
}

func withTraceComponent(trace Trace, component byte) Trace {
	trace.Channel = replaceComponentSuffix(trace.Channel, component)
	trace.Seismogram.SourceID = replaceComponentSuffix(trace.Seismogram.SourceID, component)
	for index := range trace.Seismogram.Segments {
		trace.Seismogram.Segments[index].SourceID = replaceComponentSuffix(trace.Seismogram.Segments[index].SourceID, component)
	}
	return trace
}

func replaceComponentSuffix(value string, component byte) string {
	if value == "" {
		return ""
	}
	bytes := []byte(value)
	bytes[len(bytes)-1] = component
	return string(bytes)
}

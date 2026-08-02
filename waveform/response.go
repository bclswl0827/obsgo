package waveform

import (
	"github.com/bclswl0827/obsgo/inventory"
	"github.com/bclswl0827/obsgo/signal"
)

// RemoveResponse removes an instrument response independently for every
// segment. The inventory stream supplies the response and stream-level gain.
func (seismogram Seismogram) RemoveResponse(stream inventory.Stream, output signal.ResponseOutput, options *signal.RemoveResponseOptions) (Seismogram, error) {
	result := seismogram.Clone()
	for index := range result.Segments {
		segment := &result.Segments[index]
		data, err := signal.RemoveResponse(segment.Data, segment.SampleRate, stream, output, options)
		if err != nil {
			return Seismogram{}, err
		}
		segment.Data = data
		segment.Unit = responseOutputUnit(output)
	}
	result.Unit = responseOutputUnit(output)
	return result, nil
}

// SimulateResponse simulates an instrument response independently for every
// segment. The segment boundaries and sample rates are preserved.
func (seismogram Seismogram) SimulateResponse(response *inventory.Response, input signal.ResponseOutput, options *signal.SimulateResponseOptions) (Seismogram, error) {
	result := seismogram.Clone()
	unit := simulatedResponseUnit(response)
	for index := range result.Segments {
		segment := &result.Segments[index]
		data, err := signal.SimulateResponse(segment.Data, segment.SampleRate, response, input, options)
		if err != nil {
			return Seismogram{}, err
		}
		segment.Data = data
		segment.Unit = unit
	}
	result.Unit = unit
	return result, nil
}

func simulatedResponseUnit(response *inventory.Response) string {
	if response == nil {
		return ""
	}
	if response.Sensitivity != nil && response.Sensitivity.Output.Name != "" {
		return response.Sensitivity.Output.Name
	}
	unit := ""
	stageNumber := -1
	for _, stage := range response.Stages {
		stageUnit := ""
		switch {
		case stage.FIR != nil:
			stageUnit = stage.FIR.Output.Name
		case stage.Coefficients != nil:
			stageUnit = stage.Coefficients.Output.Name
		case stage.PolesZeros != nil:
			stageUnit = stage.PolesZeros.Output.Name
		}
		if stageUnit != "" && stage.Number >= stageNumber {
			unit = stageUnit
			stageNumber = stage.Number
		}
	}
	return unit
}

func responseOutputUnit(output signal.ResponseOutput) string {
	switch output {
	case signal.Displacement:
		return "m"
	case signal.Velocity:
		return "m/s"
	case signal.Acceleration:
		return "m/s^2"
	default:
		return string(output)
	}
}

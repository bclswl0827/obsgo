package signal

import (
	"math"

	"github.com/bclswl0827/obsgo/inventory"
)

// streamResponse returns the response used by stream-level processing. Some
// SeisComP inventories store the recorder gain on Stream rather than as a
// response stage; add only the missing gain to a copy of the response.
func streamResponse(stream inventory.Stream) *inventory.Response {
	if stream.Response == nil {
		return nil
	}
	response := *stream.Response
	response.Stages = append([]inventory.Stage(nil), stream.Response.Stages...)
	if stream.GainUnit != "" || stream.Gain != 0 {
		if response.Sensitivity == nil {
			response.Sensitivity = &inventory.InstrumentSensitivity{
				Value: stream.Gain,
				Input: inventory.Units{Name: stream.GainUnit},
			}
		}
		if responseUnitName(&response) == "" && stream.GainUnit != "" {
			response.Sensitivity.Input.Name = stream.GainUnit
		}
	}
	if stream.Gain == 0 || len(response.Stages) == 0 {
		return &response
	}
	stageGain := 1.0
	hasStageGain := false
	lastNumber := 0
	for _, stage := range response.Stages {
		if stage.Number > lastNumber {
			lastNumber = stage.Number
		}
		if stage.Gain != nil {
			stageGain *= stage.Gain.Value
			hasStageGain = true
		}
	}
	if !hasStageGain || stageGain == 0 {
		response.Stages = append(response.Stages, inventory.Stage{
			Number: lastNumber + 1,
			Gain:   &inventory.StageGain{Value: stream.Gain},
		})
		return &response
	}
	ratio := stream.Gain / stageGain
	if math.Abs(ratio-1) > 1e-12*math.Max(1, math.Abs(ratio)) {
		response.Stages = append(response.Stages, inventory.Stage{
			Number: lastNumber + 1,
			Gain:   &inventory.StageGain{Value: ratio},
		})
	}
	return &response
}

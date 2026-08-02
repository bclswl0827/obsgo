package stationxml

import "github.com/bclswl0827/obsgo/inventory"

func fromResponse(response *Response) *inventory.Response {
	if response == nil {
		return nil
	}
	converted := &inventory.Response{}
	if response.Sensitivity != nil {
		converted.Sensitivity = &inventory.InstrumentSensitivity{
			Value:  response.Sensitivity.Value,
			Input:  fromUnits(response.Sensitivity.Input),
			Output: fromUnits(response.Sensitivity.Output),
		}
		if response.Sensitivity.Frequency != 0 {
			frequency := float64(response.Sensitivity.Frequency)
			converted.Sensitivity.Frequency = &frequency
		}
	}
	for _, stage := range response.Stages {
		convertedStage := inventory.Stage{Number: stage.Number}
		if stage.PolesZeros != nil {
			polesZeros := &inventory.PolesZeros{
				Name:                   stage.PolesZeros.Name,
				ResourceID:             stage.PolesZeros.ResourceID,
				Input:                  fromUnits(stage.PolesZeros.Input),
				Output:                 fromUnits(stage.PolesZeros.Output),
				TransferFunctionType:   stage.PolesZeros.TransferFunctionType,
				NormalizationFactor:    stage.PolesZeros.NormalizationFactor,
				NormalizationFrequency: stage.PolesZeros.NormalizationFrequency,
			}
			polesZeros.Zeros = fromPoleZeros(stage.PolesZeros.Zeros)
			polesZeros.Poles = fromPoleZeros(stage.PolesZeros.Poles)
			convertedStage.PolesZeros = polesZeros
			if converted.PublicID == "" {
				converted.PublicID = polesZeros.ResourceID
				converted.Name = polesZeros.Name
			}
			if converted.Type == "" {
				converted.Type = polesZeros.TransferFunctionType
			}
			if converted.NormalizationFactor == 0 {
				converted.NormalizationFactor = polesZeros.NormalizationFactor
				converted.NormalizationFrequency = polesZeros.NormalizationFrequency
			}
		}
		if stage.Coefficients != nil {
			coefficients := &inventory.Coefficients{
				Name:                 stage.Coefficients.Name,
				ResourceID:           stage.Coefficients.ResourceID,
				Input:                fromUnits(stage.Coefficients.Input),
				Output:               fromUnits(stage.Coefficients.Output),
				TransferFunctionType: stage.Coefficients.TransferFunctionType,
			}
			for _, numerator := range stage.Coefficients.Numerators {
				coefficients.Numerators = append(coefficients.Numerators, float64(numerator))
			}
			for _, denominator := range stage.Coefficients.Denominators {
				coefficients.Denominators = append(coefficients.Denominators, float64(denominator))
			}
			convertedStage.Coefficients = coefficients
		}
		if stage.FIR != nil {
			fir := &inventory.FIR{Symmetry: stage.FIR.Symmetry, Input: fromUnits(stage.FIR.Input), Output: fromUnits(stage.FIR.Output)}
			for _, numerator := range stage.FIR.Numerators {
				fir.Numerators = append(fir.Numerators, float64(numerator))
			}
			convertedStage.FIR = fir
		}
		if stage.Decimation != nil {
			convertedStage.Decimation = &inventory.Decimation{
				InputSampleRate: stage.Decimation.InputSampleRate,
				Factor:          stage.Decimation.Factor,
				Offset:          stage.Decimation.Offset,
				Delay:           stage.Decimation.Delay,
				Correction:      stage.Decimation.Correction,
			}
		}
		if stage.Gain != nil {
			convertedStage.Gain = &inventory.StageGain{Value: stage.Gain.Value, Frequency: stage.Gain.Frequency}
			if converted.Gain == 0 {
				converted.Gain = stage.Gain.Value
				converted.GainFrequency = stage.Gain.Frequency
			}
		}
		converted.Stages = append(converted.Stages, convertedStage)
	}
	return converted
}

func toResponse(response *inventory.Response) *Response {
	if response == nil {
		return nil
	}
	converted := &Response{}
	if response.Sensitivity != nil {
		frequency := 0.0
		if response.Sensitivity.Frequency != nil {
			frequency = *response.Sensitivity.Frequency
		}
		converted.Sensitivity = &InstrumentSensitivity{
			Value:     response.Sensitivity.Value,
			Frequency: Float(frequency),
			Input:     toUnits(response.Sensitivity.Input),
			Output:    toUnits(response.Sensitivity.Output),
		}
	}
	for _, stage := range response.Stages {
		convertedStage := Stage{Number: stage.Number}
		if stage.PolesZeros != nil {
			convertedStage.PolesZeros = &PolesZeros{
				Name:                   stage.PolesZeros.Name,
				ResourceID:             stage.PolesZeros.ResourceID,
				Input:                  toUnits(stage.PolesZeros.Input),
				Output:                 toUnits(stage.PolesZeros.Output),
				TransferFunctionType:   stage.PolesZeros.TransferFunctionType,
				NormalizationFactor:    stage.PolesZeros.NormalizationFactor,
				NormalizationFrequency: stage.PolesZeros.NormalizationFrequency,
			}
			convertedStage.PolesZeros.Zeros = toPoleZeros(stage.PolesZeros.Zeros)
			convertedStage.PolesZeros.Poles = toPoleZeros(stage.PolesZeros.Poles)
		}
		if stage.Coefficients != nil {
			convertedStage.Coefficients = &Coefficients{
				Name:                 stage.Coefficients.Name,
				ResourceID:           stage.Coefficients.ResourceID,
				Input:                toUnits(stage.Coefficients.Input),
				Output:               toUnits(stage.Coefficients.Output),
				TransferFunctionType: stage.Coefficients.TransferFunctionType,
			}
			for _, numerator := range stage.Coefficients.Numerators {
				convertedStage.Coefficients.Numerators = append(convertedStage.Coefficients.Numerators, Float(numerator))
			}
			for _, denominator := range stage.Coefficients.Denominators {
				convertedStage.Coefficients.Denominators = append(convertedStage.Coefficients.Denominators, Float(denominator))
			}
		}
		if stage.FIR != nil {
			convertedStage.FIR = &FIR{Symmetry: stage.FIR.Symmetry, Input: toUnits(stage.FIR.Input), Output: toUnits(stage.FIR.Output)}
			for _, numerator := range stage.FIR.Numerators {
				convertedStage.FIR.Numerators = append(convertedStage.FIR.Numerators, Float(numerator))
			}
		}
		if stage.Decimation != nil {
			convertedStage.Decimation = &Decimation{
				InputSampleRate: stage.Decimation.InputSampleRate,
				Factor:          stage.Decimation.Factor,
				Offset:          stage.Decimation.Offset,
				Delay:           stage.Decimation.Delay,
				Correction:      stage.Decimation.Correction,
			}
		}
		if stage.Gain != nil {
			convertedStage.Gain = &StageGain{Value: stage.Gain.Value, Frequency: stage.Gain.Frequency}
		}
		converted.Stages = append(converted.Stages, convertedStage)
	}
	return converted
}

func addSensor(value *inventory.Inventory, sensor *Sensor) {
	if sensor == nil || sensor.ResourceID == "" {
		return
	}
	for _, item := range value.Sensors {
		if item.PublicID == sensor.ResourceID {
			return
		}
	}
	value.Sensors = append(value.Sensors, inventory.Sensor{PublicID: sensor.ResourceID, Model: sensor.Model, Remark: sensor.Description})
}

func addDatalogger(value *inventory.Inventory, datalogger *DataLogger, channel Channel) {
	if datalogger == nil || datalogger.ResourceID == "" {
		return
	}
	for _, item := range value.Dataloggers {
		if item.PublicID == datalogger.ResourceID {
			return
		}
	}
	item := inventory.Datalogger{PublicID: datalogger.ResourceID}
	if channel.SampleRateRatio != nil {
		item.Decimation = inventory.SampleRateRatio{Numerator: channel.SampleRateRatio.NumberSamples, Denominator: channel.SampleRateRatio.NumberSeconds}
	}
	if channel.Response != nil {
		for _, stage := range channel.Response.Stages {
			if stage.Coefficients != nil {
				item.Name = stage.Coefficients.Name
			}
			if stage.Gain != nil && stage.Number > 1 {
				item.Gain = stage.Gain.Value
			}
		}
	}
	value.Dataloggers = append(value.Dataloggers, item)
}

func addResponse(value *inventory.Inventory, response *inventory.Response) {
	if response == nil || response.PublicID == "" {
		return
	}
	for _, item := range value.Responses {
		if item.PublicID == response.PublicID {
			return
		}
	}
	value.Responses = append(value.Responses, *response)
}

func fromUnits(units Units) inventory.Units {
	return inventory.Units{Name: units.Name, Description: units.Description}
}

func toUnits(units inventory.Units) Units {
	return Units{Name: units.Name, Description: units.Description}
}

func fromPoleZeros(values []PoleZero) []inventory.PoleZero {
	converted := make([]inventory.PoleZero, 0, len(values))
	for _, value := range values {
		converted = append(converted, inventory.PoleZero{Number: value.Number, Real: float64(value.Real), Imaginary: float64(value.Imaginary)})
	}
	return converted
}

func toPoleZeros(values []inventory.PoleZero) []PoleZero {
	converted := make([]PoleZero, 0, len(values))
	for _, value := range values {
		converted = append(converted, PoleZero{Number: value.Number, Real: Float(value.Real), Imaginary: Float(value.Imaginary)})
	}
	return converted
}

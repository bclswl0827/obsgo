package stationxml

import "github.com/bclswl0827/obsgo/inventory"

func fromChannel(channel Channel) inventory.Stream {
	stream := inventory.Stream{
		Code:       channel.Code,
		Start:      channel.StartDate,
		End:        channel.EndDate,
		SampleRate: channel.SampleRate,
		Depth:      channel.Depth,
		Azimuth:    channel.Azimuth,
		Dip:        channel.Dip,
		ClockDrift: float64(channel.ClockDrift),
		Format:     channel.Format,
	}
	if channel.SampleRateRatio != nil {
		stream.SampleRateNumerator = channel.SampleRateRatio.NumberSamples
		stream.SampleRateDenominator = channel.SampleRateRatio.NumberSeconds
	}
	if channel.Sensor != nil {
		stream.Sensor = channel.Sensor.ResourceID
	}
	if channel.DataLogger != nil {
		stream.Datalogger = channel.DataLogger.ResourceID
	}
	if channel.Response != nil {
		stream.Response = fromResponse(channel.Response)
	}
	if channel.Response != nil && channel.Response.Sensitivity != nil {
		stream.Gain = channel.Response.Sensitivity.Value
		stream.GainFrequency = float64(channel.Response.Sensitivity.Frequency)
		stream.GainUnit = channel.Response.Sensitivity.Input.Name
	}
	return stream
}

func toChannel(stream inventory.Stream, location inventory.SensorLocation, sensors map[string]inventory.Sensor, dataloggers map[string]inventory.Datalogger, responses map[string]inventory.Response) Channel {
	sampleRate := stream.SampleRate
	if sampleRate == 0 && stream.SampleRateDenominator != 0 {
		sampleRate = float64(stream.SampleRateNumerator) / float64(stream.SampleRateDenominator)
	}
	channel := Channel{
		Code:         stream.Code,
		LocationCode: location.Code,
		StartDate:    stream.Start,
		EndDate:      stream.End,
		Format:       stream.Format,
		Latitude:     location.Latitude,
		Longitude:    location.Longitude,
		Elevation:    location.Elevation,
		Depth:        stream.Depth,
		Azimuth:      stream.Azimuth,
		Dip:          stream.Dip,
		SampleRate:   sampleRate,
		ClockDrift:   Float(stream.ClockDrift),
	}
	if stream.SampleRateNumerator != 0 || stream.SampleRateDenominator != 0 {
		channel.SampleRateRatio = &SampleRateRatio{
			NumberSamples: stream.SampleRateNumerator,
			NumberSeconds: stream.SampleRateDenominator,
		}
	}
	if sensor, ok := sensors[stream.Sensor]; ok {
		channel.Sensor = &Sensor{ResourceID: sensor.PublicID, Description: sensor.Remark, Model: sensor.Model}
	} else if stream.Sensor != "" {
		channel.Sensor = &Sensor{ResourceID: stream.Sensor}
	}
	if datalogger, ok := dataloggers[stream.Datalogger]; ok {
		channel.DataLogger = &DataLogger{ResourceID: datalogger.PublicID}
	} else if stream.Datalogger != "" {
		channel.DataLogger = &DataLogger{ResourceID: stream.Datalogger}
	}
	if stream.Response != nil {
		channel.Response = toResponse(stream.Response)
	} else if sensor, ok := sensors[stream.Sensor]; ok {
		if response, ok := responses[sensor.Response]; ok {
			channel.Response = toResponse(&response)
		}
	}
	if datalogger, ok := dataloggers[stream.Datalogger]; ok {
		if channel.Response == nil {
			channel.Response = &Response{}
		}
		if !hasCoefficientsStage(channel.Response) {
			channel.Response.Stages = append(channel.Response.Stages, Stage{
				Number: len(channel.Response.Stages) + 1,
				Coefficients: &Coefficients{
					Name:                 datalogger.Name,
					ResourceID:           datalogger.PublicID,
					Input:                Units{Name: "V"},
					Output:               Units{Name: "COUNTS"},
					TransferFunctionType: "DIGITAL",
				},
				Decimation: &Decimation{InputSampleRate: sampleRate, Factor: 1},
				Gain:       &StageGain{Value: datalogger.Gain},
			})
		}
	}
	if stream.Gain != 0 || stream.GainFrequency != 0 || stream.GainUnit != "" {
		if channel.Response == nil {
			channel.Response = &Response{}
		}
		if channel.Response.Sensitivity == nil {
			channel.Response.Sensitivity = &InstrumentSensitivity{}
		}
		channel.Response.Sensitivity.Value = stream.Gain
		channel.Response.Sensitivity.Frequency = Float(stream.GainFrequency)
		channel.Response.Sensitivity.Input.Name = stream.GainUnit
	}
	return channel
}

func hasCoefficientsStage(response *Response) bool {
	for _, stage := range response.Stages {
		if stage.Coefficients != nil {
			return true
		}
	}
	return false
}

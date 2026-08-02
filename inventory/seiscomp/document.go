package seiscomp

import (
	"fmt"

	"github.com/bclswl0827/obsgo/inventory"
)

func (document *Document) ToInventory() (*inventory.Inventory, error) {
	if document == nil {
		return nil, fmt.Errorf("decode SeisComP inventory: document is nil")
	}
	value := &inventory.Inventory{Version: document.Version}
	for _, sensor := range document.Inventory.Sensors {
		value.Sensors = append(value.Sensors, inventory.Sensor{
			PublicID: sensor.PublicID,
			Name:     sensor.Name,
			Model:    sensor.Model,
			Unit:     sensor.Unit,
			Remark:   sensor.Remark,
			Response: sensor.Response,
		})
	}
	for _, datalogger := range document.Inventory.Dataloggers {
		converted := inventory.Datalogger{
			PublicID:             datalogger.PublicID,
			Name:                 datalogger.Name,
			RecorderModel:        datalogger.RecorderModel,
			RecorderManufacturer: datalogger.RecorderManufacturer,
			Gain:                 datalogger.Gain,
			MaxClockDrift:        datalogger.MaxClockDrift,
		}
		if datalogger.Decimation != nil {
			converted.Decimation = inventory.SampleRateRatio{
				Numerator:   datalogger.Decimation.SampleRateNumerator,
				Denominator: datalogger.Decimation.SampleRateDenominator,
			}
		}
		value.Dataloggers = append(value.Dataloggers, converted)
	}
	for _, response := range document.Inventory.Responses {
		converted, err := responseToInventory(response)
		if err != nil {
			return nil, err
		}
		value.Responses = append(value.Responses, *converted)
	}
	for _, network := range document.Inventory.Networks {
		convertedNetwork := inventory.Network{
			PublicID:     network.PublicID,
			Code:         network.Code,
			Start:        network.Start,
			End:          network.End,
			Description:  network.Description,
			Institutions: network.Institutions,
			Type:         network.Type,
		}
		for _, station := range network.Stations {
			convertedStation := inventory.Station{
				PublicID:    station.PublicID,
				Code:        station.Code,
				Start:       station.Start,
				End:         station.End,
				Description: station.Description,
				Latitude:    station.Latitude,
				Longitude:   station.Longitude,
				Elevation:   station.Elevation,
				Place:       station.Place,
				Country:     station.Country,
				Affiliation: station.Affiliation,
				Type:        station.Type,
			}
			for _, location := range station.Locations {
				convertedLocation := inventory.SensorLocation{
					PublicID:  location.PublicID,
					Code:      location.Code,
					Start:     location.Start,
					End:       location.End,
					Latitude:  location.Latitude,
					Longitude: location.Longitude,
					Elevation: location.Elevation,
				}
				for _, stream := range location.Streams {
					convertedStream := inventory.Stream{
						PublicID:              stream.PublicID,
						Code:                  stream.Code,
						Start:                 stream.Start,
						End:                   stream.End,
						Datalogger:            stream.Datalogger,
						Sensor:                stream.Sensor,
						DataloggerChannel:     stream.DataloggerChannel,
						SensorChannel:         stream.SensorChannel,
						SampleRateNumerator:   stream.SampleRateNumerator,
						SampleRateDenominator: stream.SampleRateDenominator,
						Depth:                 stream.Depth,
						Azimuth:               stream.Azimuth,
						Dip:                   stream.Dip,
						Gain:                  stream.Gain,
						GainFrequency:         stream.GainFrequency,
						GainUnit:              stream.GainUnit,
					}
					if sensor, ok := findSensor(value.Sensors, stream.Sensor); ok {
						if response, ok := findResponse(value.Responses, sensor.Response); ok {
							convertedStream.Response = response
						}
					}
					convertedLocation.Streams = append(convertedLocation.Streams, convertedStream)
				}
				convertedStation.SensorLocations = append(convertedStation.SensorLocations, convertedLocation)
			}
			convertedNetwork.Stations = append(convertedNetwork.Stations, convertedStation)
		}
		value.Networks = append(value.Networks, convertedNetwork)
	}
	return value, nil
}

func FromInventory(value *inventory.Inventory) (*Document, error) {
	if value == nil {
		return nil, fmt.Errorf("encode SeisComP inventory: inventory is nil")
	}
	document := NewDocument()
	document.Version = value.Version
	sensorResponses := make(map[string]string)
	for _, network := range value.Networks {
		for _, station := range network.Stations {
			for _, location := range station.SensorLocations {
				for _, stream := range location.Streams {
					if stream.Sensor != "" && stream.Response != nil && stream.Response.PublicID != "" {
						sensorResponses[stream.Sensor] = stream.Response.PublicID
					}
				}
			}
		}
	}
	for _, sensor := range value.Sensors {
		response := sensor.Response
		if response == "" {
			response = sensorResponses[sensor.PublicID]
		}
		document.Inventory.Sensors = append(document.Inventory.Sensors, Sensor{
			PublicID: sensor.PublicID,
			Name:     sensor.Name,
			Response: response,
			Model:    sensor.Model,
			Unit:     sensor.Unit,
			Remark:   sensor.Remark,
		})
	}
	for _, datalogger := range value.Dataloggers {
		converted := Datalogger{
			PublicID:             datalogger.PublicID,
			Name:                 datalogger.Name,
			RecorderModel:        datalogger.RecorderModel,
			RecorderManufacturer: datalogger.RecorderManufacturer,
			Gain:                 datalogger.Gain,
			MaxClockDrift:        datalogger.MaxClockDrift,
		}
		if datalogger.Decimation.Numerator != 0 || datalogger.Decimation.Denominator != 0 {
			converted.Decimation = &Decimation{
				SampleRateNumerator:   datalogger.Decimation.Numerator,
				SampleRateDenominator: datalogger.Decimation.Denominator,
			}
		}
		document.Inventory.Dataloggers = append(document.Inventory.Dataloggers, converted)
	}
	for _, response := range value.Responses {
		document.Inventory.Responses = append(document.Inventory.Responses, responseFromInventory(response))
	}
	for _, network := range value.Networks {
		convertedNetwork := Network{
			PublicID:     network.PublicID,
			Code:         network.Code,
			Start:        network.Start,
			End:          network.End,
			Description:  network.Description,
			Institutions: network.Institutions,
			Type:         network.Type,
		}
		for _, station := range network.Stations {
			convertedStation := Station{
				PublicID:    station.PublicID,
				Code:        station.Code,
				Start:       station.Start,
				End:         station.End,
				Description: station.Description,
				Latitude:    station.Latitude,
				Longitude:   station.Longitude,
				Elevation:   station.Elevation,
				Place:       station.Place,
				Country:     station.Country,
				Affiliation: station.Affiliation,
				Type:        station.Type,
			}
			for _, location := range station.SensorLocations {
				convertedLocation := SensorLocation{
					PublicID:  location.PublicID,
					Code:      location.Code,
					Start:     location.Start,
					End:       location.End,
					Latitude:  location.Latitude,
					Longitude: location.Longitude,
					Elevation: location.Elevation,
				}
				for _, stream := range location.Streams {
					convertedLocation.Streams = append(convertedLocation.Streams, Stream{
						PublicID:              stream.PublicID,
						Code:                  stream.Code,
						Start:                 stream.Start,
						End:                   stream.End,
						Datalogger:            stream.Datalogger,
						Sensor:                stream.Sensor,
						DataloggerChannel:     stream.DataloggerChannel,
						SensorChannel:         stream.SensorChannel,
						SampleRateNumerator:   stream.SampleRateNumerator,
						SampleRateDenominator: stream.SampleRateDenominator,
						Depth:                 stream.Depth,
						Azimuth:               stream.Azimuth,
						Dip:                   stream.Dip,
						Gain:                  stream.Gain,
						GainFrequency:         stream.GainFrequency,
						GainUnit:              stream.GainUnit,
					})
				}
				convertedStation.Locations = append(convertedStation.Locations, convertedLocation)
			}
			convertedNetwork.Stations = append(convertedNetwork.Stations, convertedStation)
		}
		document.Inventory.Networks = append(document.Inventory.Networks, convertedNetwork)
	}
	return document, nil
}

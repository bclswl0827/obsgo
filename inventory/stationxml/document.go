package stationxml

import (
	"fmt"

	"github.com/bclswl0827/obsgo/inventory"
)

func (document *Document) ToInventory() *inventory.Inventory {
	if document == nil {
		return nil
	}
	value := &inventory.Inventory{
		Source:    document.Source,
		Sender:    document.Sender,
		Module:    document.Module,
		ModuleURI: document.ModuleURI,
		Created:   document.Created,
	}

	for _, network := range document.Networks {
		convertedNetwork := inventory.Network{
			Code:        network.Code,
			Start:       network.StartDate,
			End:         network.EndDate,
			Description: network.Description,
		}
		for _, station := range network.Stations {
			convertedStation := inventory.Station{
				Code:         station.Code,
				Start:        station.StartDate,
				End:          station.EndDate,
				CreationDate: station.CreationDate,
				Latitude:     station.Latitude,
				Longitude:    station.Longitude,
				Elevation:    station.Elevation,
				Site: inventory.Site{
					Name:    station.Site.Name,
					Town:    station.Site.Town,
					Country: station.Site.Country,
				},
			}
			locations := make(map[string]int)
			for _, channel := range station.Channels {
				locationIndex, exists := locations[channel.LocationCode]
				if !exists {
					locationIndex = len(convertedStation.SensorLocations)
					locations[channel.LocationCode] = locationIndex
					convertedStation.SensorLocations = append(convertedStation.SensorLocations, inventory.SensorLocation{
						Code:      channel.LocationCode,
						Latitude:  channel.Latitude,
						Longitude: channel.Longitude,
						Elevation: channel.Elevation,
					})
				}
				stream := fromChannel(channel)
				convertedStation.SensorLocations[locationIndex].Streams = append(
					convertedStation.SensorLocations[locationIndex].Streams,
					stream,
				)
				addSensor(value, channel.Sensor)
				addDatalogger(value, channel.DataLogger, channel)
				addResponse(value, stream.Response)
			}
			convertedNetwork.Stations = append(convertedNetwork.Stations, convertedStation)
		}
		value.Networks = append(value.Networks, convertedNetwork)
	}
	return value
}

func FromInventory(value *inventory.Inventory) (*Document, error) {
	if value == nil {
		return nil, fmt.Errorf("encode StationXML: inventory is nil")
	}
	document := NewDocument()
	document.Source = value.Source
	document.Sender = value.Sender
	document.Module = value.Module
	document.ModuleURI = value.ModuleURI
	document.Created = value.Created
	document.Networks = make([]Network, 0, len(value.Networks))

	sensors := make(map[string]inventory.Sensor, len(value.Sensors))
	for _, sensor := range value.Sensors {
		sensors[sensor.PublicID] = sensor
	}
	dataloggers := make(map[string]inventory.Datalogger, len(value.Dataloggers))
	for _, datalogger := range value.Dataloggers {
		dataloggers[datalogger.PublicID] = datalogger
	}
	responses := make(map[string]inventory.Response, len(value.Responses))
	for _, response := range value.Responses {
		responses[response.PublicID] = response
	}

	for _, network := range value.Networks {
		convertedNetwork := Network{
			Code:        network.Code,
			StartDate:   network.Start,
			EndDate:     network.End,
			Description: network.Description,
		}
		for _, station := range network.Stations {
			convertedStation := Station{
				Code:         station.Code,
				StartDate:    station.Start,
				EndDate:      station.End,
				Latitude:     station.Latitude,
				Longitude:    station.Longitude,
				Elevation:    station.Elevation,
				CreationDate: station.CreationDate,
				Site: Site{
					Name:    station.Site.Name,
					Town:    station.Site.Town,
					Country: station.Site.Country,
				},
			}
			for _, location := range station.SensorLocations {
				for _, stream := range location.Streams {
					channel := toChannel(stream, location, sensors, dataloggers, responses)
					convertedStation.Channels = append(convertedStation.Channels, channel)
				}
			}
			convertedNetwork.Stations = append(convertedNetwork.Stations, convertedStation)
		}
		document.Networks = append(document.Networks, convertedNetwork)
	}
	return document, nil
}

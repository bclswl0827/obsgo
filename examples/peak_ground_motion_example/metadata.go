package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/bclswl0827/obsgo/inventory"
	"github.com/bclswl0827/obsgo/inventory/seiscomp"
	"github.com/bclswl0827/obsgo/inventory/stationxml"
)

func readInventory(path, format string) (*inventory.Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read inventory %q: %w", path, err)
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "auto" {
		lower := bytes.ToLower(data)
		switch {
		case bytes.Contains(lower, []byte("fdsnstationxml")):
			format = "stationxml"
		case bytes.Contains(lower, []byte("seiscomp")):
			format = "seiscomp"
		default:
			return nil, fmt.Errorf("cannot detect inventory format for %q", path)
		}
	}
	var codec inventory.Codec
	switch format {
	case "seiscomp":
		codec = seiscomp.NewCodec()
	case "stationxml":
		codec = stationxml.NewCodec()
	default:
		return nil, fmt.Errorf("unsupported inventory format %q", format)
	}
	value, err := codec.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode inventory %q: %w", path, err)
	}
	return value, nil
}

func findStream(metadata *inventory.Inventory, value trace) (inventory.Stream, error) {
	for networkIndex := range metadata.Networks {
		network := &metadata.Networks[networkIndex]
		if network.Code != value.network {
			continue
		}
		for stationIndex := range network.Stations {
			station := &network.Stations[stationIndex]
			if station.Code != value.station {
				continue
			}
			for locationIndex := range station.SensorLocations {
				location := &station.SensorLocations[locationIndex]
				if location.Code != value.location {
					continue
				}
				for streamIndex := range location.Streams {
					stream := location.Streams[streamIndex]
					if stream.Code == value.channel {
						if stream.Response == nil {
							return inventory.Stream{}, fmt.Errorf("stream %s.%s.%s.%s has no response", value.network, value.station, value.location, value.channel)
						}
						return stream, nil
					}
				}
			}
		}
	}
	return inventory.Stream{}, fmt.Errorf("stream %s.%s.%s.%s is missing from inventory", value.network, value.station, value.location, value.channel)
}

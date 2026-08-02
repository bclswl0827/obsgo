package seiscomp

import "github.com/bclswl0827/obsgo/inventory"

func findSensor(values []inventory.Sensor, publicID string) (inventory.Sensor, bool) {
	for _, value := range values {
		if value.PublicID == publicID {
			return value, true
		}
	}
	return inventory.Sensor{}, false
}

func findResponse(values []inventory.Response, publicID string) (*inventory.Response, bool) {
	for index := range values {
		if values[index].PublicID == publicID {
			return &values[index], true
		}
	}
	return nil, false
}

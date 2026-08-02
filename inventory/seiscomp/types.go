package seiscomp

import "encoding/xml"

const Namespace = "http://geofon.gfz-potsdam.de/ns/seiscomp3-schema/0.10"

type Document struct {
	XMLName   xml.Name       `xml:"seiscomp"`
	XMLNS     string         `xml:"xmlns,attr"`
	Version   string         `xml:"version,attr"`
	Inventory InventoryBlock `xml:"Inventory"`
}

type InventoryBlock struct {
	Sensors     []Sensor      `xml:"sensor,omitempty"`
	Dataloggers []Datalogger  `xml:"datalogger,omitempty"`
	Responses   []ResponsePAZ `xml:"responsePAZ,omitempty"`
	Networks    []Network     `xml:"network,omitempty"`
}

type Sensor struct {
	PublicID string `xml:"publicID,attr"`
	Name     string `xml:"name,attr,omitempty"`
	Response string `xml:"response,attr,omitempty"`
	Model    string `xml:"model,omitempty"`
	Unit     string `xml:"unit,omitempty"`
	Remark   string `xml:"remark,omitempty"`
}

type Datalogger struct {
	PublicID             string      `xml:"publicID,attr"`
	Name                 string      `xml:"name,attr,omitempty"`
	RecorderModel        string      `xml:"recorderModel,omitempty"`
	RecorderManufacturer string      `xml:"recorderManufacturer,omitempty"`
	Gain                 float64     `xml:"gain,omitempty"`
	MaxClockDrift        float64     `xml:"maxClockDrift,omitempty"`
	Decimation           *Decimation `xml:"decimation,omitempty"`
}

type Decimation struct {
	SampleRateNumerator   int `xml:"sampleRateNumerator,attr"`
	SampleRateDenominator int `xml:"sampleRateDenominator,attr"`
}

type ResponsePAZ struct {
	PublicID               string  `xml:"publicID,attr"`
	Name                   string  `xml:"name,attr,omitempty"`
	Type                   string  `xml:"type,omitempty"`
	Gain                   float64 `xml:"gain,omitempty"`
	GainFrequency          float64 `xml:"gainFrequency,omitempty"`
	NormalizationFactor    float64 `xml:"normalizationFactor,omitempty"`
	NormalizationFrequency float64 `xml:"normalizationFrequency,omitempty"`
	NumberOfZeros          int     `xml:"numberOfZeros,omitempty"`
	NumberOfPoles          int     `xml:"numberOfPoles,omitempty"`
	Zeros                  string  `xml:"zeros,omitempty"`
	Poles                  string  `xml:"poles,omitempty"`
}

type Network struct {
	PublicID     string    `xml:"publicID,attr"`
	Code         string    `xml:"code,attr"`
	Start        string    `xml:"start,omitempty"`
	End          string    `xml:"end,omitempty"`
	Description  string    `xml:"description,omitempty"`
	Institutions string    `xml:"institutions,omitempty"`
	Type         string    `xml:"type,omitempty"`
	Stations     []Station `xml:"station,omitempty"`
}

type Station struct {
	PublicID    string           `xml:"publicID,attr"`
	Code        string           `xml:"code,attr"`
	Start       string           `xml:"start,omitempty"`
	End         string           `xml:"end,omitempty"`
	Description string           `xml:"description,omitempty"`
	Latitude    float64          `xml:"latitude,omitempty"`
	Longitude   float64          `xml:"longitude,omitempty"`
	Elevation   float64          `xml:"elevation,omitempty"`
	Place       string           `xml:"place,omitempty"`
	Country     string           `xml:"country,omitempty"`
	Affiliation string           `xml:"affiliation,omitempty"`
	Type        string           `xml:"type,omitempty"`
	Locations   []SensorLocation `xml:"sensorLocation,omitempty"`
}

type SensorLocation struct {
	PublicID  string   `xml:"publicID,attr"`
	Code      string   `xml:"code,attr"`
	Start     string   `xml:"start,omitempty"`
	End       string   `xml:"end,omitempty"`
	Latitude  float64  `xml:"latitude,omitempty"`
	Longitude float64  `xml:"longitude,omitempty"`
	Elevation float64  `xml:"elevation,omitempty"`
	Streams   []Stream `xml:"stream,omitempty"`
}

type Stream struct {
	PublicID              string  `xml:"publicID,attr"`
	Code                  string  `xml:"code,attr"`
	Start                 string  `xml:"start,omitempty"`
	End                   string  `xml:"end,omitempty"`
	Datalogger            string  `xml:"datalogger,attr,omitempty"`
	Sensor                string  `xml:"sensor,attr,omitempty"`
	DataloggerChannel     int     `xml:"dataloggerChannel,omitempty"`
	SensorChannel         int     `xml:"sensorChannel,omitempty"`
	SampleRateNumerator   int     `xml:"sampleRateNumerator,omitempty"`
	SampleRateDenominator int     `xml:"sampleRateDenominator,omitempty"`
	Depth                 float64 `xml:"depth,omitempty"`
	Azimuth               float64 `xml:"azimuth,omitempty"`
	Dip                   float64 `xml:"dip,omitempty"`
	Gain                  float64 `xml:"gain,omitempty"`
	GainFrequency         float64 `xml:"gainFrequency,omitempty"`
	GainUnit              string  `xml:"gainUnit,omitempty"`
}

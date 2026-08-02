package stationxml

import (
	"encoding/xml"
	"strconv"
	"strings"
)

const (
	Namespace            = "http://www.fdsn.org/xml/station/1"
	DefaultSchemaVersion = "1.2"
)

// Float accepts the literal "None" emitted by some StationXML producers.
// StationXML normally contains numeric text, but the sample inventory uses
// this value for an unavailable instrument-sensitivity frequency.
type Float float64

func (f *Float) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var text string
	if err := decoder.DecodeElement(&text, &start); err != nil {
		return err
	}
	text = strings.TrimSpace(text)
	if text == "" || strings.EqualFold(text, "none") {
		*f = 0
		return nil
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return err
	}
	*f = Float(value)
	return nil
}

type Document struct {
	XMLName       xml.Name  `xml:"FDSNStationXML"`
	XMLNS         string    `xml:"xmlns,attr"`
	SchemaVersion string    `xml:"schemaVersion,attr"`
	Source        string    `xml:"Source"`
	Sender        string    `xml:"Sender,omitempty"`
	Module        string    `xml:"Module,omitempty"`
	ModuleURI     string    `xml:"ModuleURI,omitempty"`
	Created       string    `xml:"Created"`
	Networks      []Network `xml:"Network"`
}

type Network struct {
	Code             string    `xml:"code,attr"`
	StartDate        string    `xml:"startDate,attr,omitempty"`
	EndDate          string    `xml:"endDate,attr,omitempty"`
	RestrictedStatus string    `xml:"restrictedStatus,attr,omitempty"`
	Description      string    `xml:"Description,omitempty"`
	Stations         []Station `xml:"Station,omitempty"`
}

type Station struct {
	Code             string    `xml:"code,attr"`
	StartDate        string    `xml:"startDate,attr,omitempty"`
	EndDate          string    `xml:"endDate,attr,omitempty"`
	RestrictedStatus string    `xml:"restrictedStatus,attr,omitempty"`
	Latitude         float64   `xml:"Latitude"`
	Longitude        float64   `xml:"Longitude"`
	Elevation        float64   `xml:"Elevation"`
	Site             Site      `xml:"Site"`
	CreationDate     string    `xml:"CreationDate,omitempty"`
	Channels         []Channel `xml:"Channel,omitempty"`
}

type Site struct {
	Name    string `xml:"Name"`
	Town    string `xml:"Town,omitempty"`
	Country string `xml:"Country,omitempty"`
}

type Channel struct {
	Code             string            `xml:"code,attr"`
	LocationCode     string            `xml:"locationCode,attr"`
	StartDate        string            `xml:"startDate,attr,omitempty"`
	EndDate          string            `xml:"endDate,attr,omitempty"`
	RestrictedStatus string            `xml:"restrictedStatus,attr,omitempty"`
	Format           string            `xml:"format,omitempty"`
	Latitude         float64           `xml:"Latitude"`
	Longitude        float64           `xml:"Longitude"`
	Elevation        float64           `xml:"Elevation"`
	Depth            float64           `xml:"Depth"`
	Azimuth          float64           `xml:"Azimuth"`
	Dip              float64           `xml:"Dip"`
	SampleRate       float64           `xml:"SampleRate"`
	SampleRateRatio  *SampleRateRatio  `xml:"SampleRateRatio,omitempty"`
	ClockDrift       Float             `xml:"ClockDrift"`
	Sensor           *Sensor           `xml:"Sensor,omitempty"`
	DataLogger       *DataLogger       `xml:"DataLogger,omitempty"`
	Availability     *DataAvailability `xml:"DataAvailability,omitempty"`
	Response         *Response         `xml:"Response,omitempty"`
}

type SampleRateRatio struct {
	NumberSamples int `xml:"NumberSamples"`
	NumberSeconds int `xml:"NumberSeconds"`
}

type Sensor struct {
	ResourceID  string `xml:"resourceId,attr"`
	Description string `xml:"Description,omitempty"`
	Model       string `xml:"Model,omitempty"`
}

type DataLogger struct {
	ResourceID string `xml:"resourceId,attr"`
}

type DataAvailability struct {
	Extent Extent `xml:"Extent"`
}

type Extent struct {
	Start string `xml:"start,attr"`
	End   string `xml:"end,attr"`
}

type Response struct {
	Sensitivity *InstrumentSensitivity `xml:"InstrumentSensitivity,omitempty"`
	Stages      []Stage                `xml:"Stage,omitempty"`
}

type InstrumentSensitivity struct {
	Value     float64 `xml:"Value"`
	Frequency Float   `xml:"Frequency"`
	Input     Units   `xml:"InputUnits"`
	Output    Units   `xml:"OutputUnits"`
}

type Units struct {
	Name        string `xml:"Name"`
	Description string `xml:"Description,omitempty"`
}

type Stage struct {
	Number       int           `xml:"number,attr"`
	PolesZeros   *PolesZeros   `xml:"PolesZeros,omitempty"`
	Coefficients *Coefficients `xml:"Coefficients,omitempty"`
	FIR          *FIR          `xml:"FIR,omitempty"`
	Decimation   *Decimation   `xml:"Decimation,omitempty"`
	Gain         *StageGain    `xml:"StageGain,omitempty"`
}

type PolesZeros struct {
	Name                   string     `xml:"name,attr,omitempty"`
	ResourceID             string     `xml:"resourceId,attr,omitempty"`
	Input                  Units      `xml:"InputUnits"`
	Output                 Units      `xml:"OutputUnits"`
	TransferFunctionType   string     `xml:"PzTransferFunctionType"`
	NormalizationFactor    float64    `xml:"NormalizationFactor"`
	NormalizationFrequency float64    `xml:"NormalizationFrequency"`
	Zeros                  []PoleZero `xml:"Zero,omitempty"`
	Poles                  []PoleZero `xml:"Pole,omitempty"`
}

type PoleZero struct {
	Number    int   `xml:"number,attr"`
	Real      Float `xml:"Real"`
	Imaginary Float `xml:"Imaginary"`
}

type Coefficients struct {
	Name                 string  `xml:"name,attr,omitempty"`
	ResourceID           string  `xml:"resourceId,attr,omitempty"`
	Input                Units   `xml:"InputUnits"`
	Output               Units   `xml:"OutputUnits"`
	TransferFunctionType string  `xml:"CfTransferFunctionType"`
	Numerators           []Float `xml:"Numerator,omitempty"`
	Denominators         []Float `xml:"Denominator,omitempty"`
}

type FIR struct {
	Symmetry   string  `xml:"symmetry,attr,omitempty"`
	Input      Units   `xml:"InputUnits"`
	Output     Units   `xml:"OutputUnits"`
	Numerators []Float `xml:"NumeratorCoefficient,omitempty"`
}

type Decimation struct {
	InputSampleRate float64 `xml:"InputSampleRate"`
	Factor          int     `xml:"Factor"`
	Offset          int     `xml:"Offset"`
	Delay           float64 `xml:"Delay"`
	Correction      float64 `xml:"Correction"`
}

type StageGain struct {
	Value     float64 `xml:"Value"`
	Frequency float64 `xml:"Frequency"`
}

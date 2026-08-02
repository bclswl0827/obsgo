package inventory

import (
	"bytes"
	"io"
)

// Codec reads and writes a station inventory using one concrete XML format.
// Implementations use Inventory as the format-neutral interchange model.
type Codec interface {
	Decode(io.Reader) (*Inventory, error)
	Encode(io.Writer, *Inventory) error
}

// Transcode decodes an inventory with source and encodes it with target.
func Transcode(source, target Codec, reader io.Reader, writer io.Writer) error {
	value, err := source.Decode(reader)
	if err != nil {
		return err
	}
	return target.Encode(writer, value)
}

// TranscodeBytes converts an inventory between codecs without exposing the
// intermediate model to the caller.
func TranscodeBytes(source, target Codec, data []byte) ([]byte, error) {
	var output bytes.Buffer
	if err := Transcode(source, target, bytes.NewReader(data), &output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// Inventory is the common representation shared by the supported inventory
// formats. Date values are kept as strings so that their original precision
// and timezone representation can be retained during conversion.
type Inventory struct {
	Version     string
	Source      string
	Sender      string
	Module      string
	ModuleURI   string
	Created     string
	Sensors     []Sensor
	Dataloggers []Datalogger
	Responses   []Response
	Networks    []Network
}

type Sensor struct {
	PublicID string
	Name     string
	Model    string
	Unit     string
	Remark   string
	Response string
}

type Datalogger struct {
	PublicID             string
	Name                 string
	RecorderModel        string
	RecorderManufacturer string
	Gain                 float64
	MaxClockDrift        float64
	Decimation           SampleRateRatio
}

type Response struct {
	PublicID               string
	Name                   string
	Type                   string
	Sensitivity            *InstrumentSensitivity
	Gain                   float64
	GainFrequency          float64
	NormalizationFactor    float64
	NormalizationFrequency float64
	Zeros                  []PoleZero
	Poles                  []PoleZero
	Stages                 []Stage
}

type InstrumentSensitivity struct {
	Value     float64
	Frequency *float64
	Input     Units
	Output    Units
}

type Network struct {
	PublicID     string
	Code         string
	Start        string
	End          string
	Description  string
	Institutions string
	Type         string
	Stations     []Station
}

type Station struct {
	PublicID        string
	Code            string
	Start           string
	End             string
	CreationDate    string
	Description     string
	Latitude        float64
	Longitude       float64
	Elevation       float64
	Place           string
	Country         string
	Affiliation     string
	Type            string
	Site            Site
	SensorLocations []SensorLocation
}

type Site struct {
	Name    string
	Town    string
	Country string
}

type SensorLocation struct {
	PublicID  string
	Code      string
	Start     string
	End       string
	Latitude  float64
	Longitude float64
	Elevation float64
	Streams   []Stream
}

type Stream struct {
	PublicID              string
	Code                  string
	Start                 string
	End                   string
	Datalogger            string
	Sensor                string
	DataloggerChannel     int
	SensorChannel         int
	SampleRate            float64
	SampleRateNumerator   int
	SampleRateDenominator int
	Depth                 float64
	Azimuth               float64
	Dip                   float64
	Gain                  float64
	GainFrequency         float64
	GainUnit              string
	ClockDrift            float64
	Format                string
	Response              *Response
}

type SampleRateRatio struct {
	Numerator   int
	Denominator int
}

type Units struct {
	Name        string
	Description string
}

type Stage struct {
	Number       int
	PolesZeros   *PolesZeros
	Coefficients *Coefficients
	FIR          *FIR
	Decimation   *Decimation
	Gain         *StageGain
}

type PolesZeros struct {
	Name                   string
	ResourceID             string
	Input                  Units
	Output                 Units
	TransferFunctionType   string
	NormalizationFactor    float64
	NormalizationFrequency float64
	Zeros                  []PoleZero
	Poles                  []PoleZero
}

type PoleZero struct {
	Number    int
	Real      float64
	Imaginary float64
}

type Coefficients struct {
	Name                 string
	ResourceID           string
	Input                Units
	Output               Units
	TransferFunctionType string
	Numerators           []float64
	Denominators         []float64
}

type FIR struct {
	Symmetry   string
	Input      Units
	Output     Units
	Numerators []float64
}

type Decimation struct {
	InputSampleRate float64
	Factor          int
	Offset          int
	Delay           float64
	Correction      float64
}

type StageGain struct {
	Value     float64
	Frequency float64
}

package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/bclswl0827/obsgo/signal"
	"github.com/bclswl0827/obsgo/waveform"
	"github.com/bclswl0827/obsgo/waveform/mseed"
)

type configuration struct {
	east        string
	north       string
	vertical    string
	backAzimuth float64
	method      string
}

func main() {
	config := configuration{}
	flag.StringVar(&config.east, "east", "./examples/_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHE.D.mseed", "east-component MiniSEED file")
	flag.StringVar(&config.north, "north", "./examples/_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHN.D.mseed", "north-component MiniSEED file")
	flag.StringVar(&config.vertical, "vertical", "./examples/_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHZ.D.mseed", "vertical-component MiniSEED file")
	flag.Float64Var(&config.backAzimuth, "back-azimuth", 135, "back azimuth in degrees")
	flag.StringVar(&config.method, "method", "flinn", "polarization method: flinn or pm")
	flag.Parse()

	east, err := readTrace(config.east)
	if err != nil {
		log.Fatal(err)
	}
	north, err := readTrace(config.north)
	if err != nil {
		log.Fatal(err)
	}
	vertical, err := readTrace(config.vertical)
	if err != nil {
		log.Fatal(err)
	}

	rotatedZ, radial, transverse, err := waveform.RotateTraceZNEToZRT(vertical, north, east, config.backAzimuth)
	if err != nil {
		log.Fatalf("rotate ZNE to ZRT: %v", err)
	}
	fmt.Printf(">>> Three-component rotation\n")
	fmt.Printf("Input: %s, %s, %s\n", vertical.SourceID(), north.SourceID(), east.SourceID())
	fmt.Printf("Output: %s, %s, %s\n", rotatedZ.SourceID(), radial.SourceID(), transverse.SourceID())
	fmt.Printf("Back azimuth: %.3f°\n", config.backAzimuth)

	if len(vertical.Segments) == 0 {
		log.Fatal("vertical trace has no segments")
	}
	windowLength := int(math.Round(vertical.Segments[0].SampleRate))
	if windowLength < 2 {
		windowLength = 2
	}
	if available := len(vertical.Segments[0].Data); available < windowLength {
		windowLength = available
	}
	if windowLength < 2 {
		log.Fatal("the input is too short for polarization analysis")
	}

	method, err := polarizationMethod(config.method)
	if err != nil {
		log.Fatal(err)
	}
	polarization, err := vertical.Polarization(north, east, signal.PolarizationOptions{
		WindowLength:  windowLength,
		Step:          windowLength / 2,
		TaperFraction: 0.22,
		Method:        method,
	})
	if err != nil {
		log.Fatalf("polarization: %v", err)
	}
	if len(polarization) == 0 {
		fmt.Println("No complete polarization windows")
		return
	}
	first := polarization[0]
	value := first.Value
	fmt.Printf(">>> Polarization (%s)\n", method)
	fmt.Printf("Window center: %s\n", first.CenterTime.Format("2006-01-02T15:04:05.999999999Z07:00"))
	fmt.Printf("Azimuth: %.3f°, incidence: %.3f°\n", value.Result.Azimuth, value.Result.Incidence)
	fmt.Printf("Rectilinearity: %.6g, planarity: %.6g, ellipticity: %.6g\n",
		value.Result.Rectilinearity, value.Result.Planarity, value.Result.Ellipticity)
	fmt.Printf("Analyzed windows: %d\n", len(polarization))
}

func readTrace(path string) (waveform.Trace, error) {
	stream, err := mseed.Read(path)
	if err != nil {
		return waveform.Trace{}, err
	}
	if stream.Len() != 1 {
		return waveform.Trace{}, fmt.Errorf("MiniSEED %q contains %d traces", path, stream.Len())
	}
	return stream.Traces[0], nil
}

func polarizationMethod(value string) (signal.PolarizationMethod, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(signal.PolarizationFlinn):
		return signal.PolarizationFlinn, nil
	case string(signal.PolarizationPM):
		return signal.PolarizationPM, nil
	default:
		return "", fmt.Errorf("unsupported polarization method %q", value)
	}
}

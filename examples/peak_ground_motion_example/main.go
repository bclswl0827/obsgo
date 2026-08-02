package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bclswl0827/obsgo/signal"
)

type configuration struct {
	inventoryPath   string
	inventoryFormat string
	eastPath        string
	northPath       string
	verticalPath    string
	obsPyFFT        bool
}

func main() {
	config := parseFlags()
	if err := run(config); err != nil {
		log.Fatal(err)
	}
}

func parseFlags() configuration {
	var config configuration
	flag.StringVar(&config.inventoryPath, "inventory", "./examples/_data/seiscomp.xml", "path to SeisComP XML or StationXML metadata")
	flag.StringVar(&config.inventoryFormat, "format", "auto", "inventory format: auto, seiscomp, or stationxml")
	flag.StringVar(&config.eastPath, "east", "", "path to the east-component miniSEED file")
	flag.StringVar(&config.northPath, "north", "", "path to the north-component miniSEED file")
	flag.StringVar(&config.verticalPath, "vertical", "", "path to the vertical-component miniSEED file")
	flag.BoolVar(&config.obsPyFFT, "obspy-fft", false, "use ObsPy-compatible arbitrary FFT lengths (slower)")
	flag.Parse()
	return config
}

func run(config configuration) error {
	if config.eastPath == "" || config.northPath == "" || config.verticalPath == "" {
		return fmt.Errorf("-east, -north, and -vertical are required")
	}
	metadata, err := readInventory(config.inventoryPath, config.inventoryFormat)
	if err != nil {
		return err
	}
	traces := make([]trace, 3)
	for index, path := range []string{config.eastPath, config.northPath, config.verticalPath} {
		traces[index], err = readTrace(path)
		if err != nil {
			return err
		}
	}
	if err := validateComponents(traces); err != nil {
		return err
	}
	for index := range traces {
		traces[index], err = preprocessTrace(traces[index])
		if err != nil {
			return fmt.Errorf("preprocess %s: %w", traces[index].channel, err)
		}
	}
	fftMode := signal.FastResponseFFT
	if config.obsPyFFT {
		fftMode = signal.ObsPyResponseFFT
	}

	outputs := []struct {
		unit  signal.ResponseOutput
		label string
	}{
		{unit: signal.Acceleration, label: "PGA  (cm/s/s)"},
		{unit: signal.Velocity, label: "PGV  (cm/s)"},
		{unit: signal.Displacement, label: "PGD  (cm)"},
	}
	outputUnits := make([]signal.ResponseOutput, len(outputs))
	components := make(map[signal.ResponseOutput][][]float64, len(outputs))
	for index, output := range outputs {
		outputUnits[index] = output.unit
		components[output.unit] = make([][]float64, len(traces))
	}
	for componentIndex, currentTrace := range traces {
		stream, err := findStream(metadata, currentTrace)
		if err != nil {
			return err
		}
		processed, err := processTraceOutputs(currentTrace, stream, outputUnits, fftMode)
		if err != nil {
			return fmt.Errorf("process %s: %w", currentTrace.channel, err)
		}
		for _, output := range outputs {
			components[output.unit][componentIndex] = processed[output.unit]
		}
	}
	results := make([]float64, len(outputs))
	for outputIndex, output := range outputs {
		results[outputIndex], err = peakVectorMagnitude(components[output.unit])
		if err != nil {
			return fmt.Errorf("calculate %s: %w", output.unit, err)
		}
	}

	fmt.Println(">>> Peak Ground Motion (3-Component Vector Sum)")
	for index, output := range outputs {
		fmt.Printf("%s: %.4f\n", output.label, results[index])
	}
	return nil
}

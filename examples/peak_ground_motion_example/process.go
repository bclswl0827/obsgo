package main

import (
	"fmt"
	"math"

	"github.com/bclswl0827/obsgo/inventory"
	"github.com/bclswl0827/obsgo/signal"
)

const (
	lowCornerFrequency  = 0.1
	highCornerFrequency = 10.0
	filterOrder         = 4
	taperWidth          = 0.05
	centimetersPerMeter = 100.0
)

func preprocessTrace(value trace) (trace, error) {
	processed, err := signal.RemoveTrend(value.data, value.sampleRate)
	if err != nil {
		return trace{}, err
	}
	processed = signal.RemoveMean(processed)
	processed, err = signal.TraceCosineTaper(processed, taperWidth)
	if err != nil {
		return trace{}, err
	}
	value.data = processed
	return value, nil
}

func processTraceOutputs(value trace, stream inventory.Stream, outputs []signal.ResponseOutput, fftMode signal.ResponseFFTMode) (map[signal.ResponseOutput][]float64, error) {
	// ObsPy remove_response defaults to zero_mean=true and applies another
	// five-percent taper before entering the frequency domain.
	processed := signal.RemoveMean(value.data)
	corrected, err := signal.RemoveResponses(
		processed,
		value.sampleRate,
		stream,
		outputs,
		&signal.RemoveResponseOptions{
			WaterLevel: 60,
			TaperWidth: 0.05,
			FFTMode:    fftMode,
		},
	)
	if err != nil {
		return nil, err
	}
	for _, output := range outputs {
		corrected[output], err = signal.Bandpass(
			corrected[output],
			lowCornerFrequency,
			highCornerFrequency,
			value.sampleRate,
			filterOrder,
			false,
		)
		if err != nil {
			return nil, err
		}
		corrected[output] = signal.Multiply(corrected[output], centimetersPerMeter)
	}
	return corrected, nil
}

func peakVectorMagnitude(components [][]float64) (float64, error) {
	if len(components) != 3 {
		return 0, fmt.Errorf("exactly three processed components are required")
	}
	magnitude, err := signal.VectorMagnitude(components[0], components[1], components[2])
	if err != nil {
		return 0, err
	}
	peak := 0.0
	for _, value := range magnitude {
		peak = math.Max(peak, value)
	}
	return peak, nil
}

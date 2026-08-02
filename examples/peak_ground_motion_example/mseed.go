package main

import (
	"fmt"
	"math"
	"time"

	"github.com/bclswl0827/obsgo/waveform"
	"github.com/bclswl0827/obsgo/waveform/mseed"
)

type trace struct {
	network    string
	station    string
	location   string
	channel    string
	start      time.Time
	sampleRate float64
	data       []float64
}

func readTrace(path string) (trace, error) {
	stream, err := mseed.ReadWithOptions(path, mseed.ReadOptions{Merge: waveform.MergeOptions{
		GapPolicy: waveform.InterpolateGaps, OverlapPolicy: waveform.DropDuplicateOverlaps, ToleranceSamples: 0.5,
	}})
	if err != nil {
		return trace{}, err
	}
	if stream.Len() != 1 {
		return trace{}, fmt.Errorf("miniSEED %q contains %d channels", path, stream.Len())
	}
	value := stream.Traces[0]
	if len(value.Segments) != 1 {
		return trace{}, fmt.Errorf("miniSEED %q contains an unfilled gap", path)
	}
	segment := value.Segments[0]
	return trace{network: value.Network, station: value.Station, location: value.Location, channel: value.Channel,
		start: segment.StartTime, sampleRate: segment.SampleRate, data: segment.Data}, nil
}

func validateComponents(traces []trace) error {
	if len(traces) != 3 {
		return fmt.Errorf("exactly three components are required")
	}
	reference := traces[0]
	for _, value := range traces[1:] {
		if value.network != reference.network || value.station != reference.station || value.location != reference.location {
			return fmt.Errorf("components must belong to the same network, station, and location")
		}
		if math.Abs(value.sampleRate-reference.sampleRate)/reference.sampleRate > 1e-6 {
			return fmt.Errorf("components have different sample rates")
		}
		if math.Abs(value.start.Sub(reference.start).Seconds()*reference.sampleRate) > 0.5 {
			return fmt.Errorf("components have different start times")
		}
		if len(value.data) != len(reference.data) {
			return fmt.Errorf("components have different sample counts")
		}
	}
	return nil
}

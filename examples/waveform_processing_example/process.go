package main

import (
	"fmt"
	"math"

	"github.com/bclswl0827/obsgo/signal"
	"github.com/bclswl0827/obsgo/waveform"
)

func processTrace(trace waveform.Trace) (waveform.Trace, error) {
	if len(trace.Segments) == 0 {
		return waveform.Trace{}, fmt.Errorf("trace %s has no segments", trace.SourceID())
	}
	rate := trace.Segments[0].SampleRate
	filter, err := signal.NewButterworth(4, signal.BandPass, 0.1, 10, 1/rate)
	if err != nil {
		return waveform.Trace{}, err
	}

	result := trace.RemoveMean()
	result, err = result.RemoveTrend()
	if err != nil {
		return waveform.Trace{}, fmt.Errorf("remove trend: %w", err)
	}
	result, err = result.Taper(0.05, signal.Cosine)
	if err != nil {
		return waveform.Trace{}, fmt.Errorf("taper: %w", err)
	}
	result, err = result.Filter(filter, true)
	if err != nil {
		return waveform.Trace{}, fmt.Errorf("filter: %w", err)
	}
	return result, nil
}

func printTrace(label string, trace waveform.Trace) {
	fmt.Printf(">>> %s waveform\n", label)
	fmt.Printf("Source ID: %s\n", trace.SourceID())
	fmt.Printf("Unit: %q, segments: %d\n", trace.Unit, len(trace.Segments))
	for index, segment := range trace.Segments {
		fmt.Printf("segment %d: start=%s samples=%d rate=%.6g Hz duration=%s peak=%.6g\n",
			index+1,
			segment.StartTime.Format("2006-01-02T15:04:05.999999999Z07:00"),
			len(segment.Data),
			segment.SampleRate,
			segment.Duration(),
			peakAbs(segment.Data),
		)
	}
}

func peakAbs(data []float64) float64 {
	peak := 0.0
	for _, value := range data {
		peak = math.Max(peak, math.Abs(value))
	}
	return peak
}

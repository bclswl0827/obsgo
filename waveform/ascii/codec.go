package ascii

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bclswl0827/obsgo/waveform"
)

func Read(path string) (*waveform.Stream, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ReadReader(file)
}

func ReadReader(reader io.Reader) (*waveform.Stream, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 16*1024*1024)
	var traces []waveform.Trace
	var network, station, location, channel, sourceID, unit string
	var segments []waveform.SeismogramSegment
	var data []float64
	var start time.Time
	var rate float64
	flushSegment := func() error {
		if len(data) == 0 {
			return nil
		}
		if rate <= 0 {
			return fmt.Errorf("ASCII segment has invalid sample rate")
		}
		segments = append(segments, waveform.SeismogramSegment{Data: append([]float64(nil), data...), StartTime: start, SampleRate: rate, Unit: unit, SourceID: sourceID})
		data = nil
		return nil
	}
	flushTrace := func() error {
		if err := flushSegment(); err != nil {
			return err
		}
		if len(segments) == 0 {
			return nil
		}
		seismogram := waveform.Seismogram{Segments: segments, Unit: unit, SourceID: sourceID, SampleRate: segments[0].SampleRate, StartTime: segments[0].StartTime}
		traces = append(traces, waveform.NewTrace(network, station, location, channel, seismogram))
		segments = nil
		return nil
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if comment == "trace" || strings.HasPrefix(comment, "trace_source_id=") {
				if err := flushTrace(); err != nil {
					return nil, err
				}
			}
			if key, value, ok := strings.Cut(comment, "="); ok {
				switch strings.TrimSpace(key) {
				case "network":
					network = value
				case "station":
					station = value
				case "location":
					location = value
				case "channel":
					channel = value
				case "source_id", "trace_source_id":
					sourceID = value
				case "unit":
					unit = value
				case "segment_start":
					if err := flushSegment(); err != nil {
						return nil, err
					}
					start, _ = time.Parse(time.RFC3339Nano, value)
				case "sample_rate":
					rate, _ = strconv.ParseFloat(value, 64)
				}
			}
			continue
		}
		for _, field := range strings.Fields(line) {
			value, err := strconv.ParseFloat(field, 64)
			if err != nil {
				return nil, fmt.Errorf("parse ASCII sample %q: %w", field, err)
			}
			data = append(data, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := flushTrace(); err != nil {
		return nil, err
	}
	result := waveform.NewStream(traces...)
	if result.Len() == 0 {
		return nil, waveform.ErrEmptyData
	}
	return &result, nil
}

func Write(path string, stream waveform.Stream) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return WriteWriter(file, stream)
}

func WriteWriter(writer io.Writer, stream waveform.Stream) error {
	if stream.Len() == 0 {
		return waveform.ErrEmptyData
	}
	output := bufio.NewWriter(writer)
	if _, err := fmt.Fprintln(output, "# obsgo-ascii version=1"); err != nil {
		return err
	}
	for _, trace := range stream.Traces {
		if err := trace.Validate(); err != nil {
			return err
		}
		fmt.Fprintf(output, "# trace_source_id=%s\n# network=%s\n# station=%s\n# location=%s\n# channel=%s\n# unit=%s\n", trace.SourceID(), trace.Network, trace.Station, trace.Location, trace.Channel, trace.Unit)
		for _, segment := range trace.Segments {
			fmt.Fprintf(output, "# segment_start=%s\n# sample_rate=%s\n", segment.StartTime.Format(time.RFC3339Nano), strconv.FormatFloat(segment.SampleRate, 'g', -1, 64))
			for _, value := range segment.Data {
				fmt.Fprintln(output, strconv.FormatFloat(value, 'g', -1, 64))
			}
		}
	}
	return output.Flush()
}

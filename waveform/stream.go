package waveform

import (
	"fmt"
	"sort"
)

// Stream is a collection of channel traces, similar to ObsPy Stream.
type Stream struct {
	Traces []Trace
}

func NewStream(traces ...Trace) Stream {
	return Stream{Traces: append([]Trace(nil), traces...)}
}

func (stream Stream) Clone() Stream {
	result := Stream{Traces: make([]Trace, len(stream.Traces))}
	for index, trace := range stream.Traces {
		result.Traces[index] = trace.Clone()
	}
	return result
}

func (stream *Stream) Add(trace Trace) error {
	if err := trace.Validate(); err != nil {
		return err
	}
	stream.Traces = append(stream.Traces, trace.Clone())
	return nil
}

func (stream Stream) Len() int { return len(stream.Traces) }

func (stream Stream) At(index int) *Trace {
	if index < 0 || index >= len(stream.Traces) {
		return nil
	}
	return &stream.Traces[index]
}

func (stream Stream) Find(sourceID string) *Trace {
	for index := range stream.Traces {
		if stream.Traces[index].SourceID() == sourceID {
			return &stream.Traces[index]
		}
	}
	return nil
}

func (stream Stream) Sort() Stream {
	result := stream.Clone()
	sort.SliceStable(result.Traces, func(left, right int) bool {
		return result.Traces[left].StartTime.Before(result.Traces[right].StartTime)
	})
	return result
}

func (stream Stream) Merge(options MergeOptions) (Stream, error) {
	result := Stream{Traces: make([]Trace, len(stream.Traces))}
	for index, trace := range stream.Traces {
		merged, err := trace.Merge(options)
		if err != nil {
			return Stream{}, fmt.Errorf("merge %s: %w", trace.SourceID(), err)
		}
		result.Traces[index] = merged
	}
	return result, nil
}

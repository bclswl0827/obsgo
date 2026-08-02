package mseed

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/bclswl0827/mseedio"
	"github.com/bclswl0827/obsgo/waveform"
)

type WriteOptions struct {
	Encoding         int
	BitOrder         int
	SequenceNumber   int
	SamplesPerRecord int
}

func DefaultWriteOptions() WriteOptions {
	return WriteOptions{Encoding: mseedio.INT32, BitOrder: mseedio.MSBFIRST, SequenceNumber: 1}
}

func Write(path string, stream waveform.Stream) error {
	return WriteWithOptions(path, stream, DefaultWriteOptions())
}

func WriteWithOptions(path string, stream waveform.Stream, options WriteOptions) error {
	var data mseedio.MiniSeedData
	if err := encode(&data, stream, options); err != nil {
		return err
	}
	bytes, err := data.Encode(mseedio.OVERWRITE, options.BitOrder)
	if err != nil {
		return fmt.Errorf("encode miniSEED: %w", err)
	}
	if err := data.Write(path, mseedio.OVERWRITE, bytes); err != nil {
		return fmt.Errorf("write miniSEED %q: %w", path, err)
	}
	return nil
}

func WriteWriter(writer io.Writer, stream waveform.Stream) error {
	return WriteWriterWithOptions(writer, stream, DefaultWriteOptions())
}

func WriteWriterWithOptions(writer io.Writer, stream waveform.Stream, options WriteOptions) error {
	var data mseedio.MiniSeedData
	if err := encode(&data, stream, options); err != nil {
		return err
	}
	bytes, err := data.Encode(mseedio.OVERWRITE, options.BitOrder)
	if err != nil {
		return fmt.Errorf("encode miniSEED: %w", err)
	}
	_, err = writer.Write(bytes)
	return err
}

func encode(data *mseedio.MiniSeedData, stream waveform.Stream, options WriteOptions) error {
	if stream.Len() == 0 {
		return waveform.ErrEmptyData
	}
	if options.SequenceNumber <= 0 {
		options.SequenceNumber = 1
	}
	if err := data.Init(options.Encoding, options.BitOrder); err != nil {
		return err
	}
	sequence := options.SequenceNumber
	for _, trace := range stream.Traces {
		if err := trace.Validate(); err != nil {
			return fmt.Errorf("encode %s: %w", trace.SourceID(), err)
		}
		for _, segment := range trace.Segments {
			if options.SamplesPerRecord <= 0 {
				values, err := int32Samples(segment.Data)
				if err != nil {
					return fmt.Errorf("encode %s: %w", trace.SourceID(), err)
				}
				if err := appendRecord(data, trace, segment.StartTime, segment.SampleRate, values, sequence); err != nil {
					return err
				}
				sequence++
				continue
			}
			for start := 0; start < len(segment.Data); start += options.SamplesPerRecord {
				end := start + options.SamplesPerRecord
				if end > len(segment.Data) {
					end = len(segment.Data)
				}
				values, err := int32Samples(segment.Data[start:end])
				if err != nil {
					return fmt.Errorf("encode %s: %w", trace.SourceID(), err)
				}
				time := segment.StartTime.Add(time.Duration(float64(start) / segment.SampleRate * float64(time.Second)))
				if err := appendRecord(data, trace, time, segment.SampleRate, values, sequence); err != nil {
					return err
				}
				sequence++
			}
		}
	}
	return nil
}

func appendRecord(data *mseedio.MiniSeedData, trace waveform.Trace, start time.Time, rate float64, values []int32, sequence int) error {
	if len(values) == 0 {
		return nil
	}
	return data.Append(values, &mseedio.AppendOptions{
		SampleRate: rate, SequenceNumber: fmt.Sprintf("%06d", sequence), StartTime: start,
		NetworkCode: trace.Network, StationCode: trace.Station, LocationCode: trace.Location, ChannelCode: trace.Channel,
	})
}

func int32Samples(values []float64) ([]int32, error) {
	result := make([]int32, len(values))
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < math.MinInt32 || value > math.MaxInt32 {
			return nil, fmt.Errorf("sample %d cannot be represented as int32", index)
		}
		rounded := math.Round(value)
		if math.Abs(value-rounded) > 1e-9 {
			return nil, fmt.Errorf("sample %d is not an integer: %g", index, value)
		}
		result[index] = int32(rounded)
	}
	return result, nil
}

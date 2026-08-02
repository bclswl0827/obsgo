package mseed

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/bclswl0827/mseedio"
	"github.com/bclswl0827/obsgo/waveform"
)

type ReadOptions struct {
	Unit     string
	SourceID string
	Merge    waveform.MergeOptions
}

func DefaultReadOptions() ReadOptions {
	merge := waveform.DefaultMergeOptions()
	// Real MiniSEED headers commonly carry millisecond-quantized timestamps;
	// duplicate samples at a record boundary are therefore a normal ingest case.
	merge.OverlapPolicy = waveform.DropDuplicateOverlaps
	return ReadOptions{Merge: merge}
}

func Read(path string) (*waveform.Stream, error) {
	return ReadWithOptions(path, DefaultReadOptions())
}

func ReadWithOptions(path string, options ReadOptions) (*waveform.Stream, error) {
	var data mseedio.MiniSeedData
	if err := data.Read(path); err != nil {
		return nil, fmt.Errorf("read miniSEED %q: %w", path, err)
	}
	return decode(data, options, path)
}

func ReadReader(reader io.Reader) (*waveform.Stream, error) {
	return ReadReaderWithOptions(reader, DefaultReadOptions())
}

func ReadReaderWithOptions(reader io.Reader, options ReadOptions) (*waveform.Stream, error) {
	var data mseedio.MiniSeedData
	if err := data.ReadFromReader(reader); err != nil {
		return nil, fmt.Errorf("read miniSEED: %w", err)
	}
	return decode(data, options, "reader")
}

type record struct {
	network, station, location, channel string
	start                               time.Time
	rate                                float64
	data                                []float64
}

func decode(data mseedio.MiniSeedData, options ReadOptions, name string) (*waveform.Stream, error) {
	if len(data.Series) == 0 {
		return nil, fmt.Errorf("miniSEED %q contains no records", name)
	}
	groups := make(map[string][]record)
	keys := make([]string, 0)
	for _, series := range data.Series {
		header := series.FixedSection
		rate, err := sampleRate(header.SampleFactor, header.SampleMultiplier)
		if err != nil {
			return nil, fmt.Errorf("miniSEED %q: %w", name, err)
		}
		values, err := decodedSamples(series.DataSection.Decoded)
		if err != nil {
			return nil, fmt.Errorf("miniSEED %q: decode %s: %w", name, header.ChannelCode, err)
		}
		if len(values) != int(header.SamplesNumber) {
			return nil, fmt.Errorf("miniSEED %q: record declares %d samples, decoded %d", name, header.SamplesNumber, len(values))
		}
		network := strings.TrimSpace(header.NetworkCode)
		station := strings.TrimSpace(header.StationCode)
		location := strings.TrimSpace(header.LocationCode)
		channel := strings.TrimSpace(header.ChannelCode)
		key := strings.Join([]string{network, station, location, channel}, ".")
		if _, exists := groups[key]; !exists {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], record{
			network: network, station: station, location: location, channel: channel,
			start: header.StartTime, rate: rate, data: values,
		})
	}

	result := waveform.NewStream()
	for _, key := range keys {
		records := groups[key]
		sort.SliceStable(records, func(left, right int) bool { return records[left].start.Before(records[right].start) })
		first := records[0]
		segments := make([]waveform.SeismogramSegment, 0, len(records))
		for _, item := range records {
			segments = append(segments, waveform.SeismogramSegment{
				Data: item.data, StartTime: item.start, SampleRate: item.rate,
				Unit:     options.Unit,
				SourceID: sourceID(options.SourceID, item.network, item.station, item.location, item.channel),
			})
		}
		seismogram := waveform.Seismogram{
			Segments: segments, Unit: options.Unit, SampleRate: first.rate,
			SourceID: sourceID(options.SourceID, first.network, first.station, first.location, first.channel),
		}
		merged, err := seismogram.Merge(options.Merge)
		if err != nil {
			return nil, fmt.Errorf("merge %s: %w", first.channel, err)
		}
		if err := result.Add(waveform.NewTrace(first.network, first.station, first.location, first.channel, merged)); err != nil {
			return nil, err
		}
	}
	return &result, nil
}

func sourceID(explicit, network, station, location, channel string) string {
	if explicit != "" {
		return explicit
	}
	parts := []string{network, station, location, channel}
	for _, part := range parts {
		if part != "" {
			return strings.Join(parts, ".")
		}
	}
	return ""
}

func sampleRate(factor, multiplier int32) (float64, error) {
	if factor == 0 || multiplier == 0 {
		return 0, fmt.Errorf("sample-rate factor and multiplier must not be zero")
	}
	switch {
	case factor > 0 && multiplier > 0:
		return float64(factor) * float64(multiplier), nil
	case factor > 0:
		return float64(factor) / float64(-multiplier), nil
	case multiplier > 0:
		return float64(multiplier) / float64(-factor), nil
	default:
		return 1 / (float64(factor) * float64(multiplier)), nil
	}
}

func decodedSamples(values []any) ([]float64, error) {
	output := make([]float64, 0, len(values))
	for _, value := range values {
		switch sample := value.(type) {
		case int16:
			output = append(output, float64(sample))
		case int32:
			output = append(output, float64(sample))
		case int64:
			output = append(output, float64(sample))
		case float32:
			output = append(output, float64(sample))
		case float64:
			output = append(output, sample)
		case []int32:
			for _, item := range sample {
				output = append(output, float64(item))
			}
		case []float32:
			for _, item := range sample {
				output = append(output, float64(item))
			}
		case []float64:
			output = append(output, sample...)
		default:
			return nil, fmt.Errorf("unsupported decoded sample type %T", value)
		}
	}
	for _, sample := range output {
		if math.IsNaN(sample) || math.IsInf(sample, 0) {
			return nil, fmt.Errorf("decoded sample is not finite")
		}
	}
	return output, nil
}

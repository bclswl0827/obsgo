package waveform

import (
	"fmt"
	"math"
)

var (
	ErrSampleRateMismatch = fmt.Errorf("sample rates do not match")
	ErrUnitMismatch       = fmt.Errorf("waveform units do not match")
	ErrSourceMismatch     = fmt.Errorf("waveform source IDs do not match")
)

type GapPolicy int

const (
	PreserveGaps GapPolicy = iota
	FillGapsWithNaN
	InterpolateGaps
)

type OverlapPolicy int

const (
	RejectOverlaps OverlapPolicy = iota
	DropDuplicateOverlaps
)

type MergeOptions struct {
	GapPolicy        GapPolicy
	OverlapPolicy    OverlapPolicy
	ToleranceSamples float64
}

func DefaultMergeOptions() MergeOptions {
	return MergeOptions{GapPolicy: PreserveGaps, OverlapPolicy: RejectOverlaps, ToleranceSamples: 0.5}
}

// Merge joins adjacent segments and applies the explicit gap/overlap policy.
// The default preserves gaps and rejects overlaps, which prevents data loss.
func (seismogram Seismogram) Merge(options MergeOptions) (Seismogram, error) {
	if len(seismogram.Segments) == 0 {
		return Seismogram{}, ErrEmptyData
	}
	if options.ToleranceSamples < 0 {
		return Seismogram{}, fmt.Errorf("tolerance must not be negative")
	}
	sorted := seismogram.Sorted()
	for _, segment := range sorted.Segments {
		if err := segment.Validate(); err != nil {
			return Seismogram{}, err
		}
	}
	result := Seismogram{Unit: sorted.Unit, SourceID: sorted.SourceID, SampleRate: sorted.SampleRate}
	current := sorted.Segments[0].Clone()
	for _, incoming := range sorted.Segments[1:] {
		if !sameRate(current.SampleRate, incoming.SampleRate) {
			return Seismogram{}, ErrSampleRateMismatch
		}
		delta := incoming.StartTime.Sub(current.EndExclusive()).Seconds() * current.SampleRate
		// Record timestamps are quantized by MiniSEED headers. Interpret small
		// timestamp errors on the nearest sample grid before deciding whether
		// records are contiguous, gapped, or overlapping.
		nearest := math.Round(delta)
		if math.Abs(delta-nearest) <= options.ToleranceSamples {
			delta = nearest
		}
		switch {
		case math.Abs(delta) <= options.ToleranceSamples:
			current.Data = append(current.Data, incoming.Data...)
		case delta > 0:
			if options.GapPolicy == PreserveGaps {
				result.Segments = append(result.Segments, current)
				current = incoming.Clone()
				continue
			}
			gap := int(math.Round(delta))
			if options.GapPolicy == FillGapsWithNaN {
				current.Data = append(current.Data, make([]float64, gap)...)
				for index := len(current.Data) - gap; index < len(current.Data); index++ {
					current.Data[index] = math.NaN()
				}
			} else if options.GapPolicy == InterpolateGaps {
				left, right := current.Data[len(current.Data)-1], incoming.Data[0]
				for index := 1; index <= gap; index++ {
					fraction := float64(index) / float64(gap+1)
					current.Data = append(current.Data, left+(right-left)*fraction)
				}
			} else {
				return Seismogram{}, fmt.Errorf("unsupported gap policy %d", options.GapPolicy)
			}
			current.Data = append(current.Data, incoming.Data...)
		case delta < 0:
			if options.OverlapPolicy == RejectOverlaps {
				return Seismogram{}, ErrOverlappingData
			}
			overlap := int(math.Round(-delta))
			if overlap >= len(incoming.Data) {
				continue
			}
			current.Data = append(current.Data, incoming.Data[overlap:]...)
		}
	}
	result.Segments = append(result.Segments, current)
	if len(result.Segments) > 0 {
		result.StartTime = result.Segments[0].StartTime
	}
	return result, nil
}

func sameRate(left, right float64) bool {
	return left > 0 && right > 0 && math.Abs(left-right) <= math.Max(1e-9, math.Max(left, right)*1e-9)
}

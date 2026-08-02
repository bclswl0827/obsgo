package waveform

import (
	"math"
	"sort"
	"time"
)

// Seismogram is an ordered collection of uniformly sampled segments. Multiple
// segments represent gaps; samples are never silently fabricated by this type.
type Seismogram struct {
	Segments   []SeismogramSegment
	Unit       string
	SourceID   string
	SampleRate float64
	StartTime  time.Time
}

func NewSeismogram(segments ...SeismogramSegment) (Seismogram, error) {
	result := Seismogram{Segments: make([]SeismogramSegment, 0, len(segments))}
	for _, segment := range segments {
		if err := result.Add(segment); err != nil {
			return Seismogram{}, err
		}
	}
	return result, nil
}

func (seismogram *Seismogram) Add(segment SeismogramSegment) error {
	if err := segment.Validate(); err != nil {
		return err
	}
	if len(seismogram.Segments) > 0 {
		reference := seismogram.Segments[0]
		if !sameRate(reference.SampleRate, segment.SampleRate) {
			return ErrSampleRateMismatch
		}
		if seismogram.Unit != "" && segment.Unit != "" && seismogram.Unit != segment.Unit {
			return ErrUnitMismatch
		}
		if seismogram.SourceID != "" && segment.SourceID != "" && seismogram.SourceID != segment.SourceID {
			return ErrSourceMismatch
		}
	}
	seismogram.Segments = append(seismogram.Segments, segment.Clone())
	if seismogram.Unit == "" {
		seismogram.Unit = segment.Unit
	}
	if seismogram.SourceID == "" {
		seismogram.SourceID = segment.SourceID
	}
	if seismogram.SampleRate == 0 {
		seismogram.SampleRate = segment.SampleRate
	}
	if len(seismogram.Segments) == 1 {
		seismogram.StartTime = segment.StartTime
	}
	return nil
}

func (seismogram Seismogram) Clone() Seismogram {
	result := seismogram
	result.Segments = make([]SeismogramSegment, len(seismogram.Segments))
	for index, segment := range seismogram.Segments {
		result.Segments[index] = segment.Clone()
	}
	return result
}

func (seismogram Seismogram) Validate() error {
	if len(seismogram.Segments) == 0 {
		return ErrEmptyData
	}
	for _, segment := range seismogram.Segments {
		if err := segment.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (seismogram Seismogram) Sorted() Seismogram {
	result := seismogram.Clone()
	sort.SliceStable(result.Segments, func(left, right int) bool {
		return result.Segments[left].StartTime.Before(result.Segments[right].StartTime)
	})
	if len(result.Segments) > 0 {
		result.StartTime = result.Segments[0].StartTime
	}
	return result
}

func (seismogram Seismogram) IsContiguous(toleranceSamples float64) bool {
	if len(seismogram.Segments) < 2 {
		return true
	}
	if toleranceSamples < 0 {
		toleranceSamples = 0
	}
	sorted := seismogram.Sorted()
	for index := 1; index < len(sorted.Segments); index++ {
		previous, current := sorted.Segments[index-1], sorted.Segments[index]
		if !sameRate(previous.SampleRate, current.SampleRate) {
			return false
		}
		delta := current.StartTime.Sub(previous.EndExclusive()).Seconds() * previous.SampleRate
		if math.Abs(delta) > toleranceSamples {
			return false
		}
	}
	return true
}

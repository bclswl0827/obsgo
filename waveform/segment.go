package waveform

import (
	"fmt"
	"math"
	"time"
)

// SeismogramSegment is a uniformly sampled, contiguous block of waveform data.
// EndTime is the timestamp of the last sample, not the first instant after it.
type SeismogramSegment struct {
	Data       []float64
	StartTime  time.Time
	SampleRate float64
	Unit       string
	SourceID   string
}

// Segment is a short alias for SeismogramSegment.
type Segment = SeismogramSegment

func NewSeismogramSegment(data []float64, startTime time.Time, sampleRate float64) (SeismogramSegment, error) {
	segment := SeismogramSegment{
		Data:       append([]float64(nil), data...),
		StartTime:  startTime,
		SampleRate: sampleRate,
	}
	if err := segment.Validate(); err != nil {
		return SeismogramSegment{}, err
	}
	return segment, nil
}

func (segment SeismogramSegment) Validate() error {
	if segment.SampleRate <= 0 || math.IsNaN(segment.SampleRate) || math.IsInf(segment.SampleRate, 0) {
		return ErrInvalidSampleRate
	}
	if len(segment.Data) == 0 {
		return ErrEmptyData
	}
	return nil
}

func (segment SeismogramSegment) SamplePeriod() time.Duration {
	if segment.SampleRate <= 0 {
		return 0
	}
	return time.Duration(float64(time.Second) / segment.SampleRate)
}

func (segment SeismogramSegment) EndTime() time.Time {
	if len(segment.Data) == 0 || segment.SampleRate <= 0 {
		return segment.StartTime
	}
	return segment.StartTime.Add(time.Duration(float64(len(segment.Data)-1) / segment.SampleRate * float64(time.Second)))
}

func (segment SeismogramSegment) EndExclusive() time.Time {
	if len(segment.Data) == 0 || segment.SampleRate <= 0 {
		return segment.StartTime
	}
	return segment.StartTime.Add(time.Duration(float64(len(segment.Data)) / segment.SampleRate * float64(time.Second)))
}

func (segment SeismogramSegment) Duration() time.Duration {
	if len(segment.Data) < 2 || segment.SampleRate <= 0 {
		return 0
	}
	return segment.EndTime().Sub(segment.StartTime)
}

func (segment SeismogramSegment) Clone() SeismogramSegment {
	segment.Data = append([]float64(nil), segment.Data...)
	return segment
}

func (segment SeismogramSegment) String() string {
	return fmt.Sprintf("%s (%d samples at %g Hz)", segment.SourceID, len(segment.Data), segment.SampleRate)
}

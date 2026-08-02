package waveform

import (
	"fmt"
	"math"
	"time"

	"github.com/bclswl0827/obsgo/signal"
)

// Map applies an array-level operation independently to every segment. This
// prevents filters and calculus operations from crossing a data gap.
func (seismogram Seismogram) Map(operation func([]float64) ([]float64, error)) (Seismogram, error) {
	if operation == nil {
		return Seismogram{}, fmt.Errorf("waveform operation is nil")
	}
	result := seismogram.Clone()
	for index := range result.Segments {
		data, err := operation(result.Segments[index].Data)
		if err != nil {
			return Seismogram{}, err
		}
		result.Segments[index].Data = append([]float64(nil), data...)
	}
	return result, nil
}

func (seismogram Seismogram) RemoveMean() Seismogram {
	result := seismogram.Clone()
	for index := range result.Segments {
		result.Segments[index].Data = signal.RemoveMean(result.Segments[index].Data)
	}
	return result
}

func (seismogram Seismogram) RemoveTrend() (Seismogram, error) {
	result := seismogram.Clone()
	for index := range result.Segments {
		segment := &result.Segments[index]
		data, err := signal.RemoveTrend(segment.Data, segment.SampleRate)
		if err != nil {
			return Seismogram{}, err
		}
		segment.Data = data
	}
	return result, nil
}

func (seismogram Seismogram) Taper(width float64, taperType signal.TaperType) (Seismogram, error) {
	return seismogram.Map(func(data []float64) ([]float64, error) {
		return signal.Taper(data, width, taperType)
	})
}

func (seismogram Seismogram) Filter(filter *signal.IIRFilter, zeroPhase bool) (Seismogram, error) {
	if filter == nil {
		return Seismogram{}, fmt.Errorf("IIR filter is nil")
	}
	result := seismogram.Clone()
	for index := range result.Segments {
		filter.Reset()
		if zeroPhase {
			result.Segments[index].Data = filter.ApplyZeroPhase(result.Segments[index].Data)
		} else {
			result.Segments[index].Data = filter.Apply(result.Segments[index].Data)
		}
	}
	return result, nil
}

func (seismogram Seismogram) Differentiate() (Seismogram, error) {
	result := seismogram.Clone()
	for index := range result.Segments {
		segment := &result.Segments[index]
		data, err := signal.Differentiate(segment.Data, segment.SampleRate)
		if err != nil {
			return Seismogram{}, err
		}
		segment.Data = data
		segment.StartTime = segment.StartTime.Add(time.Duration(0.5 * float64(time.Second) / segment.SampleRate))
		segment.Unit = differentiatedUnit(segment.Unit)
	}
	if len(result.Segments) > 0 {
		result.StartTime = result.Segments[0].StartTime
	}
	result.Unit = differentiatedUnit(result.Unit)
	return result, nil
}

func (seismogram Seismogram) DifferentiateGradient() (Seismogram, error) {
	result := seismogram.Clone()
	for index := range result.Segments {
		segment := &result.Segments[index]
		data, err := signal.DifferentiateGradient(segment.Data, segment.SampleRate)
		if err != nil {
			return Seismogram{}, err
		}
		segment.Data = data
		segment.Unit = differentiatedUnit(segment.Unit)
	}
	result.Unit = differentiatedUnit(result.Unit)
	return result, nil
}

func (seismogram Seismogram) Integrate(constant float64) (Seismogram, error) {
	result := seismogram.Clone()
	prior := constant
	for index := range result.Segments {
		segment := &result.Segments[index]
		if index == 0 || !contiguousSegments(seismogram.Segments[index-1], seismogram.Segments[index]) {
			prior = constant
		}
		data, err := signal.Integrate(segment.Data, segment.SampleRate, prior)
		if err != nil {
			return Seismogram{}, err
		}
		segment.Data = data
		segment.StartTime = segment.StartTime.Add(-time.Duration(0.5 * float64(time.Second) / segment.SampleRate))
		segment.Unit = integratedUnit(segment.Unit)
		prior = data[len(data)-1]
	}
	if len(result.Segments) > 0 {
		result.StartTime = result.Segments[0].StartTime
	}
	result.Unit = integratedUnit(result.Unit)
	return result, nil
}

func contiguousSegments(left, right SeismogramSegment) bool {
	if !sameRate(left.SampleRate, right.SampleRate) {
		return false
	}
	offset := right.StartTime.Sub(left.EndExclusive()).Seconds() * left.SampleRate
	return math.Abs(offset) <= 0.25
}

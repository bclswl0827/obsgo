package waveform

import (
	"time"

	"github.com/bclswl0827/obsgo/signal"
)

// CorrelationSegment contains an ObsPy-style correlation for one aligned
// time-overlap between two waveform objects.
type CorrelationSegment struct {
	StartTime  time.Time
	SampleRate float64
	Lags       []float64
	Values     []float64
}

// Correlate correlates every time-aligned segment overlap independently.
// Gaps are never concatenated into a single artificial array.
func (seismogram Seismogram) Correlate(other Seismogram, shift int, options signal.CorrelationOptions) ([]CorrelationSegment, error) {
	pairs, err := alignPairs(seismogram, other)
	if err != nil {
		return nil, err
	}
	result := make([]CorrelationSegment, 0, len(pairs))
	for _, pair := range pairs {
		values, err := signal.Correlate(pair.Left, pair.Right, shift, options)
		if err != nil {
			return nil, err
		}
		result = append(result, CorrelationSegment{
			StartTime:  pair.StartTime,
			SampleRate: pair.SampleRate,
			Lags:       signal.CorrelationLags(len(values)),
			Values:     values,
		})
	}
	return result, nil
}

// CorrelateTemplate computes full-normalized template correlation for every
// aligned data/template segment pair.
func (seismogram Seismogram) CorrelateTemplate(template Seismogram, options signal.TemplateCorrelationOptions) ([]CorrelationSegment, error) {
	pairs, err := alignPairs(seismogram, template)
	if err != nil {
		return nil, err
	}
	result := make([]CorrelationSegment, 0, len(pairs))
	for _, pair := range pairs {
		values, err := signal.CorrelateTemplate(pair.Left, pair.Right, options)
		if err != nil {
			return nil, err
		}
		result = append(result, CorrelationSegment{
			StartTime:  pair.StartTime,
			SampleRate: pair.SampleRate,
			Lags:       signal.CorrelationLags(len(values)),
			Values:     values,
		})
	}
	return result, nil
}

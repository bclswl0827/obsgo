package waveform

import (
	"time"

	"github.com/bclswl0827/obsgo/signal"
)

// PolarizationSegment associates a windowed polarization result with the
// absolute center time of that window.
type PolarizationSegment struct {
	StartTime  time.Time
	CenterTime time.Time
	SampleRate float64
	Value      signal.PolarizationWindowResult
}

// Polarization analyzes aligned three-component waveform segments. Window
// processing never crosses a segment boundary or a time gap.
func (vertical Seismogram) Polarization(north, east Seismogram, options signal.PolarizationOptions) ([]PolarizationSegment, error) {
	segments, err := alignTriples(vertical, north, east)
	if err != nil {
		return nil, err
	}
	result := make([]PolarizationSegment, 0)
	for _, segment := range segments {
		values, err := signal.AnalyzePolarizationWindows(segment.Data[0], segment.Data[1], segment.Data[2], options)
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			center := segment.StartTime.Add(time.Duration(float64(value.CenterSample) / segment.SampleRate * float64(time.Second)))
			result = append(result, PolarizationSegment{
				StartTime:  segment.StartTime,
				CenterTime: center,
				SampleRate: segment.SampleRate,
				Value:      value,
			})
		}
	}
	return result, nil
}

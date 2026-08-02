package waveform

import (
	"fmt"
	"math"

	"github.com/bclswl0827/obsgo/signal"
)

// Resample changes the sample rate of every segment independently. Segment
// start times and metadata are preserved; no samples are created across gaps.
func (seismogram Seismogram) Resample(outputRate float64, method signal.InterpolationMethod, options signal.InterpolationOptions) (Seismogram, error) {
	return seismogram.ResampleWithOptions(outputRate, method, options)
}

// ResampleWithOptions is Resample with explicit anti-aliasing and interpolation
// controls.
func (seismogram Seismogram) ResampleWithOptions(outputRate float64, method signal.InterpolationMethod, options signal.InterpolationOptions) (Seismogram, error) {
	if outputRate <= 0 || math.IsNaN(outputRate) || math.IsInf(outputRate, 0) {
		return Seismogram{}, fmt.Errorf("output sample rate must be finite and positive")
	}
	result := seismogram.Clone()
	for index := range result.Segments {
		segment := &result.Segments[index]
		data, err := signal.ResampleTo(segment.Data, 0, segment.SampleRate, 0, outputRate,
			int(math.Round(float64(len(segment.Data)-1)*outputRate/segment.SampleRate))+1,
			method, options)
		if err != nil {
			return Seismogram{}, err
		}
		segment.Data = data
		segment.SampleRate = outputRate
	}
	result.SampleRate = outputRate
	return result, nil
}

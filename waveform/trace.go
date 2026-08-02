package waveform

import (
	"strings"
)

// Trace identifies one channel and owns its possibly discontinuous data.
type Trace struct {
	Network  string
	Station  string
	Location string
	Channel  string
	Seismogram
}

func NewTrace(network, station, location, channel string, seismogram Seismogram) Trace {
	return Trace{
		Network: network, Station: station, Location: location, Channel: channel,
		Seismogram: seismogram,
	}
}

func (trace Trace) SourceID() string {
	if trace.Seismogram.SourceID != "" {
		return trace.Seismogram.SourceID
	}
	parts := []string{trace.Network, trace.Station, trace.Location, trace.Channel}
	for _, part := range parts {
		if part != "" {
			return strings.Join(parts, ".")
		}
	}
	return ""
}

func (trace Trace) Validate() error {
	return trace.Seismogram.Validate()
}

func (trace Trace) Clone() Trace {
	trace.Seismogram = trace.Seismogram.Clone()
	return trace
}

func (trace Trace) IsContiguous(toleranceSamples float64) bool {
	return trace.Seismogram.IsContiguous(toleranceSamples)
}

func (trace Trace) Merge(options MergeOptions) (Trace, error) {
	merged, err := trace.Seismogram.Merge(options)
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = merged
	return trace, nil
}

// The Trace methods preserve channel metadata while delegating sample-level
// work to the Seismogram methods.
func (trace Trace) Map(operation func([]float64) ([]float64, error)) (Trace, error) {
	value, err := trace.Seismogram.Map(operation)
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = value
	return trace, nil
}

func (trace Trace) String() string { return trace.SourceID() }

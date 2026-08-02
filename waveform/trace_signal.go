package waveform

import (
	"github.com/bclswl0827/obsgo/inventory"
	"github.com/bclswl0827/obsgo/signal"
)

func (trace Trace) RemoveMean() Trace {
	trace.Seismogram = trace.Seismogram.RemoveMean()
	return trace
}

func (trace Trace) RemoveTrend() (Trace, error) {
	value, err := trace.Seismogram.RemoveTrend()
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = value
	return trace, nil
}

func (trace Trace) Taper(width float64, taperType signal.TaperType) (Trace, error) {
	value, err := trace.Seismogram.Taper(width, taperType)
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = value
	return trace, nil
}

func (trace Trace) Filter(filter *signal.IIRFilter, zeroPhase bool) (Trace, error) {
	value, err := trace.Seismogram.Filter(filter, zeroPhase)
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = value
	return trace, nil
}

func (trace Trace) Differentiate() (Trace, error) {
	value, err := trace.Seismogram.Differentiate()
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = value
	return trace, nil
}

func (trace Trace) DifferentiateGradient() (Trace, error) {
	value, err := trace.Seismogram.DifferentiateGradient()
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = value
	return trace, nil
}

func (trace Trace) Integrate(constant float64) (Trace, error) {
	value, err := trace.Seismogram.Integrate(constant)
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = value
	return trace, nil
}

// ObsPyHilbert preserves Trace metadata while applying the FFT Hilbert
// transform to each contiguous segment.
func (trace Trace) ObsPyHilbert() Trace {
	trace.Seismogram = trace.Seismogram.ObsPyHilbert()
	return trace
}

func (trace Trace) ObsPyEnvelope() Trace {
	trace.Seismogram = trace.Seismogram.ObsPyEnvelope()
	return trace
}

func (trace Trace) CenteredHilbert() Trace {
	trace.Seismogram = trace.Seismogram.CenteredHilbert()
	return trace
}

func (trace Trace) CenteredEnvelope() Trace {
	trace.Seismogram = trace.Seismogram.CenteredEnvelope()
	return trace
}

func (trace Trace) ComputeSpectrum(nfft int, window signal.SpectralWindow) ([]SpectrumSegment, error) {
	return trace.Seismogram.ComputeSpectrum(nfft, window)
}

func (trace Trace) PowerSpectralDensity(nfft int, window signal.SpectralWindow, options signal.SpectrumOptions) ([]PSDWaveformSegment, error) {
	return trace.Seismogram.PowerSpectralDensity(nfft, window, options)
}

func (trace Trace) Welch(options signal.WelchOptions) ([]PSDWaveformSegment, error) {
	return trace.Seismogram.Welch(options)
}

func (trace Trace) ObsPySpectrum(nfft int, window signal.SpectralWindow) ([]TwoSidedSpectrumSegment, error) {
	return trace.Seismogram.ObsPySpectrum(nfft, window)
}

func (trace Trace) ObsPyWelch(segmentLength, nfft int, overlapFraction float64, window signal.SpectralWindow) ([]TwoSidedSpectrumSegment, error) {
	return trace.Seismogram.ObsPyWelch(segmentLength, nfft, overlapFraction, window)
}

func (trace Trace) Resample(outputRate float64, method signal.InterpolationMethod, options signal.InterpolationOptions) (Trace, error) {
	value, err := trace.Seismogram.ResampleWithOptions(outputRate, method, options)
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = value
	return trace, nil
}

func (trace Trace) RemoveResponse(stream inventory.Stream, output signal.ResponseOutput, options *signal.RemoveResponseOptions) (Trace, error) {
	value, err := trace.Seismogram.RemoveResponse(stream, output, options)
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = value
	return trace, nil
}

func (trace Trace) SimulateResponse(response *inventory.Response, input signal.ResponseOutput, options *signal.SimulateResponseOptions) (Trace, error) {
	value, err := trace.Seismogram.SimulateResponse(response, input, options)
	if err != nil {
		return Trace{}, err
	}
	trace.Seismogram = value
	return trace, nil
}

func (trace Trace) Correlate(other Trace, shift int, options signal.CorrelationOptions) ([]CorrelationSegment, error) {
	return trace.Seismogram.Correlate(other.Seismogram, shift, options)
}

func (trace Trace) CorrelateTemplate(template Trace, options signal.TemplateCorrelationOptions) ([]CorrelationSegment, error) {
	return trace.Seismogram.CorrelateTemplate(template.Seismogram, options)
}

func (trace Trace) Polarization(north, east Trace, options signal.PolarizationOptions) ([]PolarizationSegment, error) {
	return trace.Seismogram.Polarization(north.Seismogram, east.Seismogram, options)
}

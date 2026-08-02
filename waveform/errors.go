package waveform

import "errors"

var (
	ErrEmptyData         = errors.New("waveform contains no samples")
	ErrInvalidSampleRate = errors.New("sample rate must be positive")
	ErrTimeDiscontinuity = errors.New("waveform contains a time discontinuity")
	ErrOverlappingData   = errors.New("waveform contains overlapping samples")
)

package sac

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/bclswl0827/obsgo/waveform"
	"github.com/bclswl0827/sacio"
)

func Read(path string) (*waveform.Stream, error) {
	var data sacio.SACData
	if err := data.Read(path); err != nil {
		return nil, fmt.Errorf("read SAC %q: %w", path, err)
	}
	if data.NPTS < 0 || int(data.NPTS) > len(data.Body) {
		return nil, fmt.Errorf("invalid SAC sample count %d", data.NPTS)
	}
	if data.DELTA <= 0 {
		return nil, fmt.Errorf("invalid SAC sample interval %g", data.DELTA)
	}
	values := make([]float64, int(data.NPTS))
	for index := range values {
		values[index] = float64(data.Body[index])
	}
	start, err := startTime(data)
	if err != nil {
		return nil, err
	}
	b := float64(data.B)
	if b < -12344 {
		b = 0
	}
	segment := waveform.SeismogramSegment{
		Data: values, StartTime: start.Add(time.Duration(b * float64(time.Second))),
		SampleRate: 1 / float64(data.DELTA), Unit: clean(string(data.KUSER0)), SourceID: sourceID(data),
	}
	trace := waveform.NewTrace(clean(string(data.KNETWK)), clean(string(data.KSTNM)), clean(string(data.KHOLE)), clean(string(data.KCMPNM)), waveform.Seismogram{
		Segments: []waveform.SeismogramSegment{segment}, SampleRate: segment.SampleRate, SourceID: segment.SourceID,
	})
	result := waveform.NewStream()
	if err := result.Add(trace); err != nil {
		return nil, err
	}
	return &result, nil
}

func Write(path string, stream waveform.Stream) error {
	if stream.Len() != 1 {
		return fmt.Errorf("SAC writer requires exactly one trace")
	}
	trace := stream.Traces[0]
	if len(trace.Segments) != 1 {
		return fmt.Errorf("SAC writer requires one continuous segment")
	}
	segment := trace.Segments[0]
	if err := segment.Validate(); err != nil {
		return err
	}
	data := sacio.SACData{}
	if err := data.Init(); err != nil {
		return err
	}
	data.DELTA = sacio.F(float32(1 / segment.SampleRate))
	data.NPTS = sacio.N(len(segment.Data))
	data.B = 0
	data.E = sacio.F(float32(segment.Duration().Seconds()))
	data.NZYEAR = sacio.N(segment.StartTime.Year())
	data.NZJDAY = sacio.N(segment.StartTime.YearDay())
	data.NZHOUR = sacio.N(segment.StartTime.Hour())
	data.NZMIN = sacio.N(segment.StartTime.Minute())
	data.NZSEC = sacio.N(segment.StartTime.Second())
	data.NZMSEC = sacio.N(segment.StartTime.Nanosecond() / int(time.Millisecond))
	data.KNETWK = sacio.K(trace.Network)
	data.KSTNM = sacio.K(trace.Station)
	data.KHOLE = sacio.K(trace.Location)
	data.KCMPNM = sacio.K(trace.Channel)
	data.KUSER0 = sacio.K(segment.Unit)
	data.IFTYPE = "itime"
	data.IZTYPE = "ib"
	data.LEVEN = true
	data.LPSPOL = true
	data.LOVROK = true
	data.LCALDA = false
	data.Body = make([]sacio.F, len(segment.Data))
	for index, value := range segment.Data {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("sample %d is not finite", index)
		}
		data.Body[index] = sacio.F(float32(value))
	}
	bytes, err := data.Encode(sacio.MSBFIRST)
	if err != nil {
		return fmt.Errorf("encode SAC: %w", err)
	}
	if err := data.Write(path, bytes); err != nil {
		return fmt.Errorf("write SAC %q: %w", path, err)
	}
	return nil
}

func sourceID(data sacio.SACData) string {
	parts := []string{clean(string(data.KNETWK)), clean(string(data.KSTNM)), clean(string(data.KHOLE)), clean(string(data.KCMPNM))}
	for _, part := range parts {
		if part != "" {
			return strings.Join(parts, ".")
		}
	}
	return ""
}

func clean(value string) string { return strings.Trim(strings.TrimSpace(value), "\x00") }

func startTime(data sacio.SACData) (time.Time, error) {
	if data.NZYEAR < 0 || data.NZJDAY < 1 || data.NZJDAY > 366 || data.NZHOUR < 0 || data.NZMIN < 0 || data.NZSEC < 0 || data.NZMSEC < 0 {
		return time.Time{}, fmt.Errorf("SAC has no valid reference time")
	}
	start := time.Date(int(data.NZYEAR), time.January, 1, int(data.NZHOUR), int(data.NZMIN), int(data.NZSEC), int(data.NZMSEC)*int(time.Millisecond), time.UTC)
	return start.AddDate(0, 0, int(data.NZJDAY)-1), nil
}

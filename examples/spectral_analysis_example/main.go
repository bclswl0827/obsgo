package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bclswl0827/obsgo/signal"
	"github.com/bclswl0827/obsgo/waveform/mseed"
)

func main() {
	input := flag.String("input", "./examples/_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHZ.D.mseed", "input MiniSEED file")
	flag.Parse()

	stream, err := mseed.Read(*input)
	if err != nil {
		log.Fatal(err)
	}
	if stream.Len() != 1 {
		log.Fatalf("expected one trace, got %d", stream.Len())
	}
	trace := stream.Traces[0]

	spectra, err := trace.ComputeSpectrum(0, signal.HannWindow)
	if err != nil {
		log.Fatalf("compute spectrum: %v", err)
	}
	psds, err := trace.PowerSpectralDensity(0, signal.HannWindow, signal.SpectrumOptions{Demean: true})
	if err != nil {
		log.Fatalf("compute PSD: %v", err)
	}
	welch, err := trace.Welch(signal.DefaultWelchOptions())
	if err != nil {
		log.Fatalf("compute Welch PSD: %v", err)
	}

	fmt.Printf(">>> Spectral analysis\n")
	fmt.Printf("Source ID: %s\n", trace.SourceID())
	fmt.Printf("Input segments: %d\n", len(trace.Segments))
	for index := range spectra {
		value := spectra[index].Value
		psd := psds[index].Value
		fmt.Printf("segment %d: nfft=%d bins=%d dominant=%.6g Hz max PSD=%.6g %s\n",
			index+1,
			value.NFFT,
			len(value.Frequencies),
			dominantFrequency(value.Frequencies, value.Amplitude),
			maximum(psd.Power),
			psds[index].Unit+"²/Hz",
		)
	}
	if len(welch) > 0 {
		value := welch[0].Value
		fmt.Printf("Welch: segments=%d nfft=%d bins=%d max PSD=%.6g %s\n",
			len(welch), value.NFFT, len(value.Frequencies), maximum(value.Power), welch[0].Unit+"²/Hz")
	}
}

func dominantFrequency(frequencies, amplitude []float64) float64 {
	if len(frequencies) == 0 || len(frequencies) != len(amplitude) {
		return 0
	}
	best := 0
	for index := 1; index < len(amplitude); index++ {
		if amplitude[index] > amplitude[best] {
			best = index
		}
	}
	return frequencies[best]
}

func maximum(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maximum := values[0]
	for _, value := range values[1:] {
		if value > maximum {
			maximum = value
		}
	}
	return maximum
}

package main

import (
	"flag"
	"log"

	"github.com/bclswl0827/obsgo/waveform/mseed"
)

type configuration struct {
	input string
}

func main() {
	config := configuration{}
	flag.StringVar(&config.input, "input", "./examples/_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHZ.D.mseed", "input MiniSEED file")
	flag.Parse()

	stream, err := mseed.Read(config.input)
	if err != nil {
		log.Fatal(err)
	}
	if stream.Len() != 1 {
		log.Fatalf("expected one trace, got %d", stream.Len())
	}

	trace := stream.Traces[0]
	printTrace("input", trace)
	processed, err := processTrace(trace)
	if err != nil {
		log.Fatal(err)
	}
	printTrace("processed", processed)
}

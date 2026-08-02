package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bclswl0827/obsgo/waveform"
	"github.com/bclswl0827/obsgo/waveform/ascii"
	"github.com/bclswl0827/obsgo/waveform/mseed"
	"github.com/bclswl0827/obsgo/waveform/sac"
)

func main() {
	input := flag.String("input", "./examples/_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHZ.D.mseed", "input MiniSEED file")
	output := flag.String("output", "./examples/waveform_io_example/output", "output directory")
	flag.Parse()

	options := mseed.DefaultReadOptions()
	options.Merge.GapPolicy = waveform.InterpolateGaps
	options.Merge.OverlapPolicy = waveform.DropDuplicateOverlaps
	stream, err := mseed.ReadWithOptions(*input, options)
	if err != nil {
		log.Fatal(err)
	}
	if stream.Len() != 1 {
		log.Fatalf("expected one trace, got %d", stream.Len())
	}
	if err := os.MkdirAll(*output, 0o755); err != nil {
		log.Fatal(err)
	}

	asciiPath := filepath.Join(*output, "trace.txt")
	sacPath := filepath.Join(*output, "trace.sac")
	if err := ascii.Write(asciiPath, *stream); err != nil {
		log.Fatalf("write ASCII: %v", err)
	}
	fromASCII, err := ascii.Read(asciiPath)
	if err != nil {
		log.Fatalf("read ASCII: %v", err)
	}
	if err := sac.Write(sacPath, *fromASCII); err != nil {
		log.Fatalf("write SAC: %v", err)
	}
	fromSAC, err := sac.Read(sacPath)
	if err != nil {
		log.Fatalf("read SAC: %v", err)
	}

	fmt.Println(">>> Waveform IO round trip")
	printSummary("MiniSEED", *stream)
	printSummary("ASCII", *fromASCII)
	printSummary("SAC", *fromSAC)
	fmt.Printf("Wrote: %s\n       %s\n", asciiPath, sacPath)
}

func printSummary(format string, stream waveform.Stream) {
	if stream.Len() == 0 {
		fmt.Printf("%s: empty\n", format)
		return
	}
	trace := stream.Traces[0]
	fmt.Printf("%s: source=%s unit=%q segments=%d samples=%d rate=%.6g Hz\n",
		format, trace.SourceID(), trace.Unit, len(trace.Segments), sampleCount(trace), trace.Segments[0].SampleRate)
}

func sampleCount(trace waveform.Trace) int {
	count := 0
	for _, segment := range trace.Segments {
		count += len(segment.Data)
	}
	return count
}

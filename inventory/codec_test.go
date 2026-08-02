package inventory_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/bclswl0827/obsgo/inventory"
	"github.com/bclswl0827/obsgo/inventory/seiscomp"
	"github.com/bclswl0827/obsgo/inventory/stationxml"
)

func TestSampleInventoriesTranscodeBothDirections(t *testing.T) {
	seiscompData, err := os.ReadFile("../examples/_data/seiscomp.xml")
	if err != nil {
		t.Fatal(err)
	}
	stationXMLData, err := os.ReadFile("../examples/_data/stationxml.xml")
	if err != nil {
		t.Fatal(err)
	}

	codecs := []struct {
		name   string
		data   []byte
		source inventory.Codec
		target inventory.Codec
	}{
		{name: "SeisComP to StationXML", data: seiscompData, source: seiscomp.NewCodec(), target: stationxml.NewCodec()},
		{name: "StationXML to SeisComP", data: stationXMLData, source: stationxml.NewCodec(), target: seiscomp.NewCodec()},
	}

	for _, testCase := range codecs {
		t.Run(testCase.name, func(t *testing.T) {
			converted, err := inventory.TranscodeBytes(testCase.source, testCase.target, testCase.data)
			if err != nil {
				t.Fatal(err)
			}
			if len(bytes.TrimSpace(converted)) == 0 {
				t.Fatal("conversion returned empty XML")
			}
			decoded, err := testCase.target.Decode(bytes.NewReader(converted))
			if err != nil {
				t.Fatal(err)
			}
			if len(decoded.Networks) != 1 {
				t.Fatalf("decoded network count = %d", len(decoded.Networks))
			}
			if len(decoded.Networks[0].Stations) != 1 {
				t.Fatalf("decoded station count = %d", len(decoded.Networks[0].Stations))
			}
		})
	}
}

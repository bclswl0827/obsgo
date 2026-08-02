package stationxml

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarshalParseAndStreamCodec(t *testing.T) {
	document := NewDocument()
	document.Source = "unit-test"
	document.Created = "2024-01-02T03:04:05Z"
	document.Networks = []Network{{Code: "XX", Stations: []Station{{
		Code: "TEST", Site: Site{Name: "Test station"}, Channels: []Channel{{
			Code: "BHZ", LocationCode: "00", Response: &Response{
				Sensitivity: &InstrumentSensitivity{Value: 2, Frequency: 1, Input: Units{Name: "m/s"}, Output: Units{Name: "count"}},
			},
		}},
	}}}}

	data, err := Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `<FDSNStationXML xmlns="`+Namespace+`"`) {
		t.Fatalf("xml = %s", data)
	}
	decoded, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.XMLNS != Namespace || decoded.SchemaVersion != DefaultSchemaVersion {
		t.Fatalf("namespace = %q, schemaVersion = %q", decoded.XMLNS, decoded.SchemaVersion)
	}
	if decoded.Networks[0].Stations[0].Channels[0].Response.Sensitivity.Value != 2 {
		t.Fatalf("decoded = %#v", decoded)
	}

	var buffer bytes.Buffer
	if err := Encode(&buffer, decoded); err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(&buffer); err != nil {
		t.Fatal(err)
	}
}

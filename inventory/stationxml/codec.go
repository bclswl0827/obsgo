package stationxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/bclswl0827/obsgo/inventory"
)

// Codec implements inventory.Codec for FDSN StationXML.
type Codec struct{}

var _ inventory.Codec = Codec{}

func NewCodec() Codec { return Codec{} }

func (Codec) Decode(reader io.Reader) (*inventory.Inventory, error) {
	document, err := decodeDocument(reader)
	if err != nil {
		return nil, err
	}
	return document.ToInventory(), nil
}

func (Codec) Encode(writer io.Writer, value *inventory.Inventory) error {
	document, err := FromInventory(value)
	if err != nil {
		return err
	}
	return Encode(writer, document)
}

// NewDocument creates an empty StationXML document with the standard
// namespace and schema version.
func NewDocument() *Document {
	return &Document{XMLNS: Namespace, SchemaVersion: DefaultSchemaVersion}
}

// Parse decodes a StationXML document while retaining its format-specific
// representation. Use Codec.Decode when the common inventory model is needed.
func Parse(data []byte) (*Document, error) { return Decode(bytes.NewReader(data)) }

func Decode(reader io.Reader) (*Document, error) { return decodeDocument(reader) }

func decodeDocument(reader io.Reader) (*Document, error) {
	var document Document
	if err := xml.NewDecoder(reader).Decode(&document); err != nil {
		return nil, fmt.Errorf("decode StationXML: %w", err)
	}
	if document.XMLNS == "" {
		document.XMLNS = document.XMLName.Space
	}
	if document.XMLNS == "" {
		document.XMLNS = Namespace
	}
	if document.SchemaVersion == "" {
		document.SchemaVersion = DefaultSchemaVersion
	}
	return &document, nil
}

// Marshal encodes a format-specific StationXML document.
func Marshal(document *Document) ([]byte, error) {
	var buffer bytes.Buffer
	if err := Encode(&buffer, document); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Encode(writer io.Writer, document *Document) error {
	if document == nil {
		return fmt.Errorf("encode StationXML: document is nil")
	}
	value := *document
	if value.XMLNS == "" {
		value.XMLNS = Namespace
	}
	if value.SchemaVersion == "" {
		value.SchemaVersion = DefaultSchemaVersion
	}
	if _, err := io.WriteString(writer, xml.Header); err != nil {
		return fmt.Errorf("encode StationXML header: %w", err)
	}
	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode StationXML: %w", err)
	}
	if err := encoder.Flush(); err != nil {
		return fmt.Errorf("encode StationXML flush: %w", err)
	}
	return nil
}

// DecodeInventory is the common-model form of Decode.
func DecodeInventory(reader io.Reader) (*inventory.Inventory, error) {
	return Codec{}.Decode(reader)
}

// EncodeInventory writes a common inventory as StationXML.
func EncodeInventory(writer io.Writer, value *inventory.Inventory) error {
	return Codec{}.Encode(writer, value)
}

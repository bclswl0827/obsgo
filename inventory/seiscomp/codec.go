package seiscomp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/bclswl0827/obsgo/inventory"
)

// Codec implements inventory.Codec for the SeisComP XML inventory format.
type Codec struct{}

var _ inventory.Codec = Codec{}

func NewCodec() Codec { return Codec{} }

func (Codec) Decode(reader io.Reader) (*inventory.Inventory, error) {
	document, err := decodeDocument(reader)
	if err != nil {
		return nil, err
	}
	return document.ToInventory()
}

func (Codec) Encode(writer io.Writer, value *inventory.Inventory) error {
	document, err := FromInventory(value)
	if err != nil {
		return err
	}
	return Encode(writer, document)
}

func NewDocument() *Document { return &Document{XMLNS: Namespace, Version: "0.10"} }

func Parse(data []byte) (*Document, error) { return Decode(bytes.NewReader(data)) }

func Decode(reader io.Reader) (*Document, error) { return decodeDocument(reader) }

func decodeDocument(reader io.Reader) (*Document, error) {
	var document Document
	if err := xml.NewDecoder(reader).Decode(&document); err != nil {
		return nil, fmt.Errorf("decode SeisComP inventory: %w", err)
	}
	if document.XMLNS == "" {
		document.XMLNS = document.XMLName.Space
	}
	if document.XMLNS == "" {
		document.XMLNS = Namespace
	}
	if document.Version == "" {
		document.Version = "0.10"
	}
	return &document, nil
}

func Marshal(document *Document) ([]byte, error) {
	var buffer bytes.Buffer
	if err := Encode(&buffer, document); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Encode(writer io.Writer, document *Document) error {
	if document == nil {
		return fmt.Errorf("encode SeisComP inventory: document is nil")
	}
	value := *document
	if value.XMLNS == "" {
		value.XMLNS = Namespace
	}
	if value.Version == "" {
		value.Version = "0.10"
	}
	if _, err := io.WriteString(writer, xml.Header); err != nil {
		return fmt.Errorf("encode SeisComP header: %w", err)
	}
	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode SeisComP inventory: %w", err)
	}
	if err := encoder.Flush(); err != nil {
		return fmt.Errorf("encode SeisComP flush: %w", err)
	}
	return nil
}

func DecodeInventory(reader io.Reader) (*inventory.Inventory, error) { return Codec{}.Decode(reader) }

func EncodeInventory(writer io.Writer, value *inventory.Inventory) error {
	return Codec{}.Encode(writer, value)
}

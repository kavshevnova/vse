package main

// Задача: JSONEncoder, XMLEncoder, MultiEncoder (CompositeEncoder).

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
)

type Encoder interface {
	Encode(data interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
	ContentType() string
}

type MultiEncoder interface {
	AddEncoder(name string, encoder Encoder)
	Encode(data interface{}) (map[string][]byte, error)
	DecodeWith(name string, data []byte, v interface{}) error
}

// --- JSONEncoder ---

type JSONEncoder struct{}

func (e *JSONEncoder) Encode(data interface{}) ([]byte, error) { return json.Marshal(data) }
func (e *JSONEncoder) Decode(data []byte, v interface{}) error { return json.Unmarshal(data, v) }
func (e *JSONEncoder) ContentType() string                     { return "application/json" }

// --- XMLEncoder ---

type XMLEncoder struct{}

func (e *XMLEncoder) Encode(data interface{}) ([]byte, error) { return xml.Marshal(data) }
func (e *XMLEncoder) Decode(data []byte, v interface{}) error { return xml.Unmarshal(data, v) }
func (e *XMLEncoder) ContentType() string                     { return "application/xml" }

// --- CompositeEncoder ---

type CompositeEncoder struct {
	encoders map[string]Encoder
}

func NewCompositeEncoder() *CompositeEncoder {
	return &CompositeEncoder{encoders: make(map[string]Encoder)}
}

func (c *CompositeEncoder) AddEncoder(name string, encoder Encoder) {
	c.encoders[name] = encoder
}

func (c *CompositeEncoder) Encode(data interface{}) (map[string][]byte, error) {
	result := make(map[string][]byte)
	for name, enc := range c.encoders {
		b, err := enc.Encode(data)
		if err != nil {
			return nil, fmt.Errorf("encoder %s: %w", name, err)
		}
		result[name] = b
	}
	return result, nil
}

func (c *CompositeEncoder) DecodeWith(name string, data []byte, v interface{}) error {
	enc, ok := c.encoders[name]
	if !ok {
		return errors.New("encoder not found: " + name)
	}
	return enc.Decode(data, v)
}

type Sample struct {
	XMLName xml.Name `xml:"sample" json:"-"`
	Name    string   `json:"name" xml:"name"`
	Value   int      `json:"value" xml:"value"`
}

func main() {
	composite := NewCompositeEncoder()
	composite.AddEncoder("json", &JSONEncoder{})
	composite.AddEncoder("xml", &XMLEncoder{})

	data := Sample{Name: "test", Value: 42}
	results, err := composite.Encode(data)
	if err != nil {
		panic(err)
	}
	for name, b := range results {
		fmt.Printf("%s: %s\n", name, string(b))
	}
}

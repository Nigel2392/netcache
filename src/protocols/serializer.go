package protocols

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
)

type Serializer interface {
	Serialize(any) ([]byte, error)
	Deserialize(any, []byte) error
}

type JsonSerializer struct{}

func (s *JsonSerializer) Serialize(a any) ([]byte, error) {
	return json.Marshal(a)
}

func (s *JsonSerializer) Deserialize(a any, b []byte) error {
	return json.Unmarshal(b, a)
}

type XmlSerializer struct{}

func (s *XmlSerializer) Serialize(a any) ([]byte, error) {
	return xml.Marshal(a)
}

func (s *XmlSerializer) Deserialize(a any, b []byte) error {
	return xml.Unmarshal(b, a)
}

type GobSerializer struct{}

func (s *GobSerializer) Serialize(a any) ([]byte, error) {
	var b = new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(a)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (s *GobSerializer) Deserialize(a any, b []byte) error {
	var buf = bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	return dec.Decode(a)
}

package protocols

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"
)

type MessageType int8

func (m MessageType) String() string {
	return msgTypeMap[m]
}

const (
	TypeSET MessageType = iota
	TypeGET
	TypeDELETE
	TypeCLEAR
	TypeHAS
	TypeKEYS
	TypeERROR
	TypeEND
)

var msgTypeMap = map[MessageType]string{
	TypeSET:    "SET",
	TypeGET:    "GET",
	TypeDELETE: "DELETE",
	TypeCLEAR:  "CLEAR",
	TypeHAS:    "HAS",
	TypeKEYS:   "KEYS",
	TypeERROR:  "ERROR",
	TypeEND:    "END",
}

// A message to be sent, or read from.
//
// It is formatted in a littleEndian binary format, with the following format:
//
// Type (int8) | TTL (int64) | Key Length (int64) | Key (string) | Value Length (int64) | Value ([]byte)
type Message struct {
	Type  MessageType
	TTL   time.Duration
	Key   string
	Value []byte
}

func WriteEnd(w io.Writer) error {
	var message = &Message{
		Type: TypeEND,
	}

	var _, err = message.WriteTo(w)
	return err
}

func (m Message) WriteTo(w io.Writer) (n int64, err error) {
	var b = new(bytes.Buffer)
	if w == nil {
		return 0, io.ErrClosedPipe
	}
	err = binary.Write(b, binary.LittleEndian, m.Type)
	if err != nil {
		return 0, err
	}
	err = binary.Write(b, binary.LittleEndian, m.TTL)
	if err != nil {
		return 0, err
	}
	err = binary.Write(b, binary.LittleEndian, int64(len(m.Key)))
	if err != nil {
		return 0, err
	}
	err = binary.Write(b, binary.LittleEndian, []byte(m.Key))
	if err != nil {
		return 0, err
	}
	err = binary.Write(b, binary.LittleEndian, int64(len(m.Value)))
	if err != nil {
		return 0, err
	}
	err = binary.Write(b, binary.LittleEndian, m.Value)
	if err != nil {
		return 0, err
	}
	err = binary.Write(w, binary.LittleEndian, int64(b.Len()))
	if err != nil {
		return 0, err
	}

	return b.WriteTo(w)
}

type messageError int

const (
	ErrInvalidFormat messageError = iota
	ErrUnexpectedEOF
)

func (m messageError) Error() string {
	return msgErrs[m]
}

var msgErrs = map[messageError]string{
	ErrInvalidFormat: "invalid format",
	ErrUnexpectedEOF: "unexpected EOF",
}

func (m messageError) Is(target error) bool {
	if target == nil {
		return false
	}
	t, ok := target.(messageError)
	if !ok {
		return false
	}
	return t == m
}

func (m *Message) ReadFrom(r io.Reader) (int64, error) {
	var (
		err error
		b   = new(bytes.Buffer)
	)

	var size int64
	err = binary.Read(r, binary.LittleEndian, &size)
	if err != nil {
		return 0, err
	}

	_, err = io.CopyN(b, r, size)
	if err != nil {
		return 0, err
	}

	err = binary.Read(b, binary.LittleEndian, &m.Type)
	if err != nil {
		return 0, err
	}

	err = binary.Read(b, binary.LittleEndian, &m.TTL)
	if err != nil {
		return 0, err
	}

	var keySize int64
	err = binary.Read(b, binary.LittleEndian, &keySize)
	if err != nil {
		return 0, err
	}

	key := make([]byte, keySize)
	_, err = io.ReadFull(b, key)
	if err != nil {
		return 0, err
	}
	m.Key = string(key)

	var valueSize int64
	err = binary.Read(b, binary.LittleEndian, &valueSize)
	if err != nil {
		return 0, err
	}

	value := make([]byte, valueSize)
	_, err = io.ReadFull(b, value)
	if err != nil {
		return 0, err
	}
	m.Value = value
	return size, nil
}

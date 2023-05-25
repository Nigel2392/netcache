package protocols_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/Nigel2392/netcache/src/protocols"
)

func TestProtocol(t *testing.T) {
	var message = &protocols.Message{
		Type:  protocols.TypeGET,
		Key:   "kgsjhdfsghjgfdey",
		Value: []byte("vallkfkngkjsdfbgvsdfgvaue"),
		TTL:   5 * time.Second,
	}

	var b bytes.Buffer
	_, err := message.WriteTo(&b)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(b.Bytes())
	fmt.Println(b.String())

	var message2 = new(protocols.Message)
	_, err = message2.ReadFrom(&b)
	if err != nil {
		t.Fatalf("error reading from buffer %s %v", err.Error(), message2)
	}

	if message.Type != message2.Type {
		t.Fatalf("type mismatch %d != %d", message.Type, message2.Type)
	}

	if message.Key != message2.Key {
		t.Fatalf("key mismatch %s != %s", message.Key, message2.Key)
	}

	if string(message.Value) != string(message2.Value) {
		t.Fatalf("value mismatch %s != %s", string(message.Value), string(message2.Value))
	}

	if message.TTL != message2.TTL {
		t.Fatalf("ttl mismatch %d != %d", message.TTL, message2.TTL)
	}

	t.Log(message2.Key)
	t.Log(string(message2.Value))
	t.Log(message2.TTL)
}

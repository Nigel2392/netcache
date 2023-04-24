package client

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Nigel2392/netcache/src/cache"
	"github.com/Nigel2392/netcache/src/protocols"
)

func init() {
	// here to make sure the cache client implements the cache interface
	var _ = Cache(&CacheClient{})
}

type cacheItem struct {
	// The value of the item.
	value interface{}

	// The expiration time of the item.
	ttl time.Duration
}

// Get the value of the item.
func (i *cacheItem) Value() interface{} {
	return i.value
}

// Get the expiration time of the item.
func (i *cacheItem) TTL() time.Duration {
	return i.ttl
}

type CacheClient struct {
	// The address of the server.
	ServerAddr string

	// The connection to the server.
	conn net.Conn

	// The serializer to use for values.
	Serializer protocols.Serializer
}

// Create a new cache client.
func New(serverAddr string, serializer protocols.Serializer) *CacheClient {
	var c = &CacheClient{
		ServerAddr: serverAddr,
	}

	if serializer != nil {
		c.Serializer = serializer
	} else {
		c.Serializer = &protocols.GobSerializer{}
	}

	return c
}

// Connect to the cache.
func (c *CacheClient) Connect() error {
	var err error
	if c.Serializer == nil {
		c.Serializer = &protocols.GobSerializer{}
	}
	c.conn, err = net.Dial("tcp", c.ServerAddr)
	return err
}

// Close the connection to the cache.
func (c *CacheClient) Close() error {
	return c.conn.Close()
}

// Get an item from the cache.
//
// Destination is only used if a serializer has been set.
func (c *CacheClient) Get(key string, dst any) (Item, error) {
	if err := cache.IsValidKey(key); err != nil {
		return nil, err
	}

	var message = &protocols.Message{
		Type: protocols.TypeGET,
		Key:  key,
	}

	fmt.Println("sending message")
	_, err := message.WriteTo(c.conn)
	if err != nil {
		return nil, err
	}

	fmt.Println("reading message")
	_, err = message.ReadFrom(c.conn)
	if err != nil {
		return nil, err
	}

	fmt.Println("got message", message)

	if message.Type == protocols.TypeERROR {
		return nil, fmt.Errorf("error from server: %s", message.Value)
	} else if message.Type != protocols.TypeGET {
		return nil, fmt.Errorf("unexpected message type from server instead of GET message: %d", message.Type)
	}

	err = c.listenForEnd()
	if err != nil {
		return nil, err
	}

	var v any
	if dst != nil && c.Serializer != nil {
		v = dst
		err = c.Serializer.Deserialize(v, message.Value)
		if err != nil {
			return nil, err
		}
	} else {
		v = message.Value
	}

	return &cacheItem{
		value: v,
		ttl:   message.TTL,
	}, nil
}

// Set an item in the cache.
//
// If a serializer has been set, the value will be serialized.
//
// Otherwise, the value must be a []byte or string.
func (c *CacheClient) Set(key string, value any, ttl time.Duration) error {
	if err := cache.IsValidKey(key); err != nil {
		return err
	}

	var v []byte
	var err error
	if c.Serializer == nil {
		val, ok := value.([]byte)
		if !ok {
			val, ok := value.(string)
			if !ok {
				return fmt.Errorf("no serializer set and value is not a []byte or string")
			}
			v = []byte(val)
		} else {
			v = val
		}
	} else {
		v, err = c.Serializer.Serialize(value)
		if err != nil {
			return err
		}
	}

	var message = &protocols.Message{
		Type:  protocols.TypeSET,
		Key:   key,
		Value: v,
		TTL:   ttl,
	}

	_, err = message.WriteTo(c.conn)
	if err != nil {
		return err
	}

	return c.listenForEnd()
}

// Delete an item from the cache.
func (c *CacheClient) Delete(key string) error {
	if err := cache.IsValidKey(key); err != nil {
		return err
	}

	var message = &protocols.Message{
		Type: protocols.TypeDELETE,
		Key:  key,
	}

	_, err := message.WriteTo(c.conn)
	if err != nil {
		return err
	}

	return c.listenForEnd()
}

// Clear the cache.
func (c *CacheClient) Clear() error {
	var message = &protocols.Message{
		Type: protocols.TypeCLEAR,
	}

	_, err := message.WriteTo(c.conn)
	if err != nil {
		return err
	}

	return c.listenForEnd()
}

// Check if the cache has an item.
func (c *CacheClient) Has(key string) (bool, error) {
	if err := cache.IsValidKey(key); err != nil {
		return false, err
	}

	var message = &protocols.Message{
		Type: protocols.TypeHAS,
		Key:  key,
	}

	_, err := message.WriteTo(c.conn)
	if err != nil {
		return false, err
	}

	_, err = message.ReadFrom(c.conn)
	if err != nil {
		return false, err
	}

	if message.Type == protocols.TypeERROR {
		return false, fmt.Errorf("error from server: %s", message.Value)
	} else if message.Type != protocols.TypeHAS {
		return false, fmt.Errorf("unexpected message type from server instead of HAS message: %d", message.Type)
	}

	var b bool
	err = c.Serializer.Deserialize(&b, message.Value)
	if err != nil {
		return false, err
	}

	err = c.listenForEnd()
	if err != nil {
		return false, err
	}

	return true, nil
}

// Retrieve the keys in the cache.
func (c *CacheClient) Keys() ([]string, error) {
	var message = &protocols.Message{
		Type: protocols.TypeKEYS,
	}

	_, err := message.WriteTo(c.conn)
	if err != nil {
		return nil, err
	}

	_, err = message.ReadFrom(c.conn)
	if err != nil {
		return nil, err
	}

	if message.Type == protocols.TypeERROR {
		return nil, fmt.Errorf("error from server: %s", message.Value)
	}
	err = c.listenForEnd()
	if err != nil {
		return nil, err
	}
	message.Value = bytes.Trim(message.Value, ",")
	var keys []string = strings.Split(string(message.Value), ",")
	for i, key := range keys {
		key = strings.TrimSpace(key)
		if err := cache.IsValidKey(key); err != nil {

		}
		keys[i] = key
	}

	return keys, nil
}

func (c *CacheClient) listenForEnd() error {
	var message = new(protocols.Message)
	_, err := message.ReadFrom(c.conn)
	if err != nil {
		return err
	}
	if message.Type == protocols.TypeERROR {
		return fmt.Errorf("error from server: %s", message.Value)
	} else if message.Type != protocols.TypeEND {
		return fmt.Errorf("unexpected message from server instead of END message: %v, %d", message, message.Type)
	}
	return nil
}

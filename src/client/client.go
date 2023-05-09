package client

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/Nigel2392/go-datastructures/stack"
	"github.com/Nigel2392/netcache/src/cache"
	"github.com/Nigel2392/netcache/src/protocols"
)

func init() {
	// here to make sure the cache client implements the cache interface
	var _ = Cache(&CacheClient{})
}

type connectionPool struct {
	// The address of the server.
	ServerAddr string

	// the connection pool
	pool *stack.Stack[net.Conn]

	// mutex
	mu sync.Mutex
}

func newPool(serverAddr string, connections int) (*connectionPool, error) {

	var p = &connectionPool{
		ServerAddr: serverAddr,
		pool:       &stack.Stack[net.Conn]{},
	}

	for i := 0; i < connections; i++ {
		var conn, err = net.Dial("tcp", serverAddr)
		if err != nil {
			return nil, err
		}
		p.pool.Push(conn)
	}

	return p, nil
}

// get a connection from the pool
func (p *connectionPool) get(deadline time.Duration) net.Conn {
	if deadline == 0 {
		deadline = 5 * time.Second
	}

	if p == nil {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	var conn, ok = p.pool.PopOK()
	if !ok {
		var c, okChan = p.pool.PopOKDeadline(deadline / 2)
		select {
		case <-okChan:
			return nil
		case conn = <-c:
			deadline = deadline / 2
		}
	}

	if err := conn.SetDeadline(time.Now().Add(deadline)); err != nil {
		return nil
	}

	return conn
}

func (p *connectionPool) put(conn net.Conn) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pool.Push(conn)
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
	pool *connectionPool

	// The serializer to use for values.
	Serializer protocols.Serializer

	timeout time.Duration

	// the amount of connections to keep open
	connections int
}

// Create a new cache client.
func New(serverAddr string, serializer protocols.Serializer, timeout time.Duration, connections int) *CacheClient {
	var c = &CacheClient{
		ServerAddr: serverAddr,
	}

	if serializer != nil {
		c.Serializer = serializer
	} else {
		c.Serializer = &protocols.GobSerializer{}
	}

	if timeout != 0 {
		c.timeout = timeout
	} else {
		c.timeout = 5 * time.Second
	}

	if connections != 0 {
		c.connections = connections
	} else {
		c.connections = 10
	}

	return c
}

// Connect to the cache.
func (c *CacheClient) Connect() error {
	var err error
	var pool *connectionPool
	pool, err = newPool(c.ServerAddr, c.connections)
	if err != nil {
		// Set the client to nil, return the error.
		var valueOf = reflect.ValueOf(c)
		valueOf.
			Elem().
			Set(
				reflect.Zero(
					valueOf.Elem().Type(),
				),
			)
		return err
	}
	c.pool = pool
	return nil
}

// Close the connection to the cache.
func (c *CacheClient) Close() error {
	if c == nil {
		return nil
	}
	for i := 0; i < c.connections; i++ {
		var conn = c.pool.get(c.timeout)
		if conn == nil {
			continue
		}
		if err := conn.Close(); err != nil {
			return err
		}

	}
	return nil
}

// Get an item from the cache.
//
// Destination is only used if a serializer has been set.
func (c *CacheClient) Get(key string, dst any) (Item, error) {
	if c == nil {
		return nil, fmt.Errorf("cache client is nil")
	}
	if err := cache.IsValidKey(key); err != nil {
		return nil, err
	}

	var message = &protocols.Message{
		Type: protocols.TypeGET,
		Key:  key,
	}

	var conn = c.pool.get(c.timeout)
	defer c.pool.put(conn)
	_, err := message.WriteTo(conn)
	if err != nil {
		return nil, err
	}

	_, err = message.ReadFrom(conn)
	if err != nil {
		return nil, err
	}

	if message.Type == protocols.TypeERROR {
		return nil, fmt.Errorf("error from server: %s", message.Value)
	} else if message.Type != protocols.TypeGET {
		return nil, fmt.Errorf("unexpected message type from server instead of GET message: %d", message.Type)
	}

	err = c.listenForEnd(conn)
	if err != nil {
		return nil, err
	}

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
	if c == nil {
		return fmt.Errorf("cache client is nil")
	}
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

	var conn = c.pool.get(c.timeout)
	defer c.pool.put(conn)
	_, err = message.WriteTo(conn)
	if err != nil {
		return err
	}

	return c.listenForEnd(conn)
}

// Delete an item from the cache.
func (c *CacheClient) Delete(key string) error {
	if c == nil {
		return fmt.Errorf("cache client is nil")
	}
	if err := cache.IsValidKey(key); err != nil {
		return err
	}

	var message = &protocols.Message{
		Type: protocols.TypeDELETE,
		Key:  key,
	}

	var conn = c.pool.get(c.timeout)
	defer c.pool.put(conn)
	var _, err = message.WriteTo(conn)
	if err != nil {
		return err
	}
	return c.listenForEnd(conn)
}

// Clear the cache.
func (c *CacheClient) Clear() error {
	if c == nil {
		return fmt.Errorf("cache client is nil")
	}
	var message = &protocols.Message{
		Type: protocols.TypeCLEAR,
	}

	var conn = c.pool.get(c.timeout)
	defer c.pool.put(conn)
	var _, err = message.WriteTo(conn)
	if err != nil {
		return err
	}
	return c.listenForEnd(conn)
}

// Check if the cache has an item.
func (c *CacheClient) Has(key string) (bool, error) {
	if c == nil {
		return false, fmt.Errorf("cache client is nil")
	}
	if err := cache.IsValidKey(key); err != nil {
		return false, err
	}

	var message = &protocols.Message{
		Type: protocols.TypeHAS,
		Key:  key,
	}
	var conn = c.pool.get(c.timeout)
	defer c.pool.put(conn)
	_, err := message.WriteTo(conn)
	if err != nil {
		return false, err
	}

	_, err = message.ReadFrom(conn)
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

	err = c.listenForEnd(conn)

	if err != nil {
		return false, err
	}

	return true, nil
}

// Retrieve the keys in the cache.
func (c *CacheClient) Keys() ([]string, error) {
	if c == nil {
		return nil, fmt.Errorf("cache client is nil")
	}
	var message = &protocols.Message{
		Type: protocols.TypeKEYS,
	}

	var conn = c.pool.get(c.timeout)
	defer c.pool.put(conn)
	_, err := message.WriteTo(conn)
	if err != nil {
		return nil, err
	}

	_, err = message.ReadFrom(conn)
	if err != nil {
		return nil, err
	}

	if message.Type == protocols.TypeERROR {
		return nil, fmt.Errorf("error from server: %s", message.Value)
	}
	err = c.listenForEnd(conn)
	if err != nil {
		return nil, err
	}

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

func (c *CacheClient) listenForEnd(conn net.Conn) error {
	var message = new(protocols.Message)
	_, err := message.ReadFrom(conn)
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

package cache

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"
)

// A simple in-memory cache implementation based on a map of string[TYPE].
type MemoryCache[T any] struct {
	cache           map[string]*memitem[T]
	cleanupInterval time.Duration
	cleanupTicker   *time.Ticker
	closed          chan struct{}
	mu              sync.Mutex
	lastTick        time.Time
}

// Returns a new in-memory cache.
func NewMemoryCache() Cache {
	return NewGenericMemoryCache[[]byte]()
}

// Dump the cache to bytes.
func (c *MemoryCache[T]) Dump() ([]byte, error) {
	var buf bytes.Buffer
	var enc = json.NewEncoder(&buf)
	c.mu.Lock()
	defer c.mu.Unlock()
	err := enc.Encode(c.cache)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Load the cache from bytes.
func (c *MemoryCache[T]) Load(data []byte) error {
	var buf = bytes.NewBuffer(data)
	var dec = json.NewDecoder(buf)
	c.mu.Lock()
	defer c.mu.Unlock()
	err := dec.Decode(&c.cache)
	if err != nil {
		return err
	}
	return nil
}

// Might as well make it generic, right?
func NewGenericMemoryCache[T any]() *MemoryCache[T] {
	return &MemoryCache[T]{
		cache:  make(map[string]*memitem[T]),
		closed: make(chan struct{}),
	}
}

func (c *MemoryCache[T]) Run(interval time.Duration) {
	c.closed = make(chan struct{})
	c.cleanupInterval = interval
	go c.work()
}

func (c *MemoryCache[T]) Set(key string, value T, ttl time.Duration) (inserted bool, err error) {
	var item *memitem[T]
	item = &memitem[T]{
		Key:   key,
		Value: value,
		TTL:   ttl,
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = item
	return true, nil
}

func (c *MemoryCache[T]) Get(key string) (value T, ttl time.Duration, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var item, ok = c.cache[key]
	if !ok {
		return value, 0, ErrItemNotFound
	}
	item.TTL -= time.Since(c.lastTick)
	return item.Value, item.TTL, nil
}

func (c *MemoryCache[T]) Delete(key string) (deleted bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var _, ok = c.cache[key]
	if !ok {
		return false, ErrItemNotFound
	}
	delete(c.cache, key)
	return true, nil
}

func (c *MemoryCache[T]) Clear() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*memitem[T])
	return nil
}

func (c *MemoryCache[T]) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var keys = make([]string, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	return keys
}

func (c *MemoryCache[T]) Close() {
	close(c.closed)
}

func (c *MemoryCache[T]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.cache)
}

func (c *MemoryCache[T]) Has(key string) (ttl time.Duration, has bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var item, ok = c.cache[key]
	if !ok {
		return 0, false
	}
	item.TTL -= time.Since(c.lastTick)
	return item.TTL, true
}

func (c *MemoryCache[T]) work() {
	c.cleanupTicker = time.NewTicker(c.cleanupInterval)
	c.lastTick = time.Now()
	for {
		select {
		case <-c.cleanupTicker.C:
			c.mu.Lock()
			for key, item := range c.cache {
				item.TTL -= time.Since(c.lastTick)
				if item.TTL <= 0 {
					delete(c.cache, key)
				}
			}
			c.lastTick = time.Now()
			c.mu.Unlock()
		case <-c.closed:
			c.cleanupTicker.Stop()
			return
		}
	}
}

package client

import "time"

// A item to be used inside of a cache.
type Item interface {
	// Get the value of the item.
	Value() interface{}
	// Get the expiration time of the item.
	TTL() time.Duration
}

// A cache to store items in.
type Cache interface {
	// Connect to the cache.
	Connect() error
	// Get an item from the cache.
	Get(key string, dst any) (Item, error)
	// Set an item in the cache.
	Set(key string, value any, ttl time.Duration) error
	// Delete an item from the cache.
	Delete(key string) error
	// Clear the cache.
	Clear() error
	// Check if the cache has an item.
	Has(key string) (bool, error)
	// Keys returns all keys in the cache.
	Keys() ([]string, error)
	// Ping the cache.
	Ping() error
}

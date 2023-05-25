package cache

import "time"

type Cache interface {
	// Initialize the cache with the set cleanup interval.
	//
	// Extra initialization can be done here.
	Run(interval time.Duration)
	// Set a value in the cache.
	Set(key string, value []byte, ttl time.Duration) (inserted bool, err error)
	// Get a value from the cache.
	Get(key string) (value []byte, ttl time.Duration, err error)
	// Delete a value from the cache.
	Delete(key string) (deleted bool, err error)
	// Clear the cache.
	Clear() (err error)
	// Keys returns all keys in the cache.
	Keys() []string
	// Close the cache.
	Close()
	// Len returns the number of items in the cache.
	Len() int
	// Has returns true if the key exists in the cache.
	Has(key string) (ttl time.Duration, has bool)

	// Dumps the cache to bytes.
	Dump() ([]byte, error)
	// Loads the cache from bytes.
	Load([]byte) error
}

package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Nigel2392/netcache/src/cache/binarytree"
)

const DefaultCleanupInterval = 5 * time.Minute

type queueItem struct {
	item  *item
	value []byte
}

// A cache.
//
// Saves items in the specified cached directory.
//
// Item keys are stored inside of a binary tree.
type FileCache struct {
	cache           binarytree.InterfacedBST[*item]
	cleanupInterval time.Duration
	cleanupTicker   *time.Ticker
	closed          chan struct{}
	dir             string
	mu              sync.Mutex
	queue           chan *queueItem
	lastTick        time.Time
}

// Create a new cache.
func NewFileCache(dir string) Cache {
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	return &FileCache{
		cache: binarytree.InterfacedBST[*item]{},
		dir:   dir,
	}
}

// Dump the cache to bytes.
func (c *FileCache) Dump() ([]byte, error) {
	var buf bytes.Buffer
	var enc = gob.NewEncoder(&buf)
	c.mu.Lock()
	defer c.mu.Unlock()
	err := enc.Encode(c.cache)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Load the cache from bytes.
func (c *FileCache) Load(data []byte) error {
	var buf = bytes.NewBuffer(data)
	var dec = gob.NewDecoder(buf)
	c.mu.Lock()
	defer c.mu.Unlock()
	err := dec.Decode(&c.cache)
	if err != nil {
		return err
	}

	// Verify the integrity of the cache.
	//
	// Delete any items not found in the filesystem.
	if c.cache.Len() > 0 {
		return c.VerifyIntegrity()
	}

	return nil
}

// Verify the integrity of the cache.
func (c *FileCache) VerifyIntegrity() error {
	var errs []error = make([]error, 0)

	c.cache.DeleteIf(func(i *item) bool {
		var _, itemPath = i.getpath(c.dir)
		var _, err = os.Stat(itemPath)
		if err != nil {
			errs = append(errs, err)
		}
		return err != nil
	})

	return NewIntegrityError(errs)
}

//
//	// Connect the cache.
//	func (c *FileCache) Connect() error {
//		select {
//		case <-c.closed:
//			if c.cleanupInterval == 0 {
//				c.cleanupInterval = DefaultCleanupInterval
//			}
//			c.Run(c.cleanupInterval)
//		default:
//			return ErrCacheAlreadyRunning
//		}
//		return nil
//	}

// Run the cache.
func (c *FileCache) Run(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queue = make(chan *queueItem, 100)
	c.closed = make(chan struct{})
	c.cleanupInterval = interval
	go c.work()
}

// Set an item in the cache.
func (c *FileCache) Set(key string, value []byte, ttl time.Duration) (inserted bool, err error) {
	var (
		item *item
	)
	item, err = newItem(key, ttl)
	if err != nil {
		return false, err
	}

	c.push(item, value)

	select {
	case err = <-item.err:
		if err != nil {
			return false, err
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	inserted = c.cache.Insert(item)
	return inserted, nil
}

// Get an item from the cache.
func (c *FileCache) Get(key string) (value []byte, ttl time.Duration, err error) {
	var itm *item
	var liveItem *item
	var found bool
	itm, err = newItemKey(key)
	c.mu.Lock()
	defer c.mu.Unlock()
	liveItem, found = c.cache.Search(itm)
	if !found {
		return nil, 0, ErrItemNotFound
	}
	value, err = liveItem.read(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			c.cache.Delete(liveItem)
			return nil, 0, ErrItemNotFound
		}
		return nil, 0, err
	}

	liveItem.TTL -= time.Since(c.lastTick)
	if liveItem.TTL <= 0 {
		c.cache.Delete(liveItem)
		liveItem.delete(c.dir)
		return nil, 0, ErrItemNotFound
	}

	return value, liveItem.TTL, nil
}

// Delete an item from the cache.
func (c *FileCache) Delete(key string) (deleted bool, err error) {
	var item *item
	item, err = newItemKey(key)
	if err != nil {
		return false, err
	}
	err = c.delete(item)
	if err != nil {
		return false, err
	}
	deleted = c.cache.Delete(item)
	if !deleted {
		return false, ErrItemNotFound
	}
	return deleted, nil
}

// Clear the cache.
func (c *FileCache) Clear() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var errors []error = make([]error, 0)
	c.cache.Traverse(func(i *item) {
		err = i.delete(c.dir)
		if err != nil {
			errors = append(errors, err)
		}
	})

	if len(errors) > 0 {
		return fmt.Errorf("%d errors have occurred trying to clear the cache", len(errors))
	}

	return nil
}

// Retrieve the keys from the cache.
func (c *FileCache) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var keys []string = make([]string, c.cache.Len())
	var i int
	c.cache.Traverse(func(item *item) {
		keys[i] = item.Key
		i++
	})
	return keys
}

// Close the cache.
func (c *FileCache) Close() {
	close(c.closed)
}

// Return the number of items in the cache.
func (c *FileCache) Len() int {
	return c.cache.Len()
}

// Check if the cache has an item.
func (c *FileCache) Has(key string) (ttl time.Duration, has bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var item *item = &item{Key: key}
	item, has = c.cache.Search(item)
	if !has {
		return 0, false
	}

	item.TTL -= time.Since(c.lastTick)
	if item.TTL <= 0 {
		c.cache.Delete(item)
		item.delete(c.dir)
		return 0, false
	}

	return item.TTL, true
}

func (c *FileCache) delete(item *item) (err error) {
	var locked bool = c.mu.TryLock()
	if locked {
		defer c.mu.Unlock()
	}

	var _, found = c.cache.Search(item)
	if !found {
		c.mu.Unlock()
		return ErrItemNotFound
	}

	err = item.delete(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			c.cache.Delete(item)
			return ErrItemNotFound
		}
		return err
	}

	return nil
}

func (c *FileCache) push(item *item, value []byte) {
	var queueItem = &queueItem{
		item:  item,
		value: value,
	}

	c.queue <- queueItem
}

func (c *FileCache) work() {
	c.cleanupTicker = time.NewTicker(c.cleanupInterval)
	c.lastTick = time.Now()
	defer c.cleanupTicker.Stop()
	for {
		select {
		case <-c.closed:
			return
		case <-c.cleanupTicker.C:
			c.mu.Lock()
			c.cleanup()
			c.lastTick = time.Now()
			c.mu.Unlock()
		case item := <-c.queue:
			item.item.write(c.dir, item.value)
		}
	}
}

func (c *FileCache) cleanup() {
	var err error
	c.cache.DeleteIf(func(i *item) bool {
		if i == nil {
			return true
		}
		i.TTL -= c.cleanupInterval
		if i.TTL <= 0 {
			err = c.delete(i)
			if err != nil {
				panic(err)
			}
			return true
		}
		return false
	})
}

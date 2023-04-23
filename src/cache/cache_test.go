package cache_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/Nigel2392/netcache/src/cache"
)

const CACHE_DIR = "./cache-tests"

type cacheItem struct {
	key   string
	value []byte
}

var cacheItems [128]*cacheItem

func init() {
	for i := 0; i < len(cacheItems); i++ {
		cacheItems[i] = &cacheItem{
			key:   "key" + strconv.Itoa(i),
			value: []byte("value" + strconv.Itoa(i)),
		}
	}
}

func TestFileCache(t *testing.T) {
	var c = cache.NewFileCache(CACHE_DIR)
	c.Run(1 * time.Second)

	for _, item := range cacheItems {
		var inserted, err = c.Set(item.key, item.value, 5*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if !inserted {
			t.Fatalf("item not inserted %s", item.key)
		}
	}

	var keys = c.Keys()
	for _, item := range cacheItems {
		var found = false
		for _, key := range keys {
			if key == item.key {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("key not found %s", item.key)
		}
	}

	for _, item := range cacheItems {
		var value, ttl, err = c.Get(item.key)
		if err != nil {
			t.Fatal(err)
		}

		if ttl <= 0 {
			t.Fatalf("ttl not positive %s", item.key)
		}

		if string(value) != string(item.value) {
			t.Fatalf("value mismatch %s != %s", string(value), string(item.value))
		}
	}

	for _, item := range cacheItems {
		if ttl, has := c.Has(item.key); !has || ttl <= 0 {
			t.Fatalf("item expired %s", item.key)
		}
	}

	for _, item := range cacheItems {
		var deleted, err = c.Delete(item.key)
		if err != nil {
			t.Fatal(err)
		}
		if !deleted {
			t.Fatalf("item not deleted %s", item.key)
		}
	}

	for _, item := range cacheItems {
		var value, ttl, err = c.Get(item.key)
		if !cache.ErrItemNotFound.Is(err) {
			t.Fatal(err)
		}
		if value != nil {
			t.Fatalf("value not nil %s", item.key)
		}
		if ttl > 0 {
			t.Fatalf("ttl not zero %s", item.key)
		}
	}
}

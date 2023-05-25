package cache

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

//	var allowedChars = map[rune]bool{
//		'.': true,
//		'-': true,
//		'_': true,
//	}
//	for _, char := range key {
//		if !allowedChars[char] && !('a' <= char && char <= 'z') && !('A' <= char && char <= 'Z') && !('0' <= char && char <= '9') {
//			return fmt.Errorf("key '%s' contains invalid characters", key)
//		}
//	}

var keyRegexFunc = regexp.MustCompile(`^[a-zA-Z0-9\._\-]+$`).MatchString

type memitem[T any] struct {
	Key   string
	Value T
	TTL   time.Duration
}

type item struct {
	Key      string        // the key the filename of the cached item, this cannot contain any special characters
	Hash     uint64        // the hash is the directory the key is stored in
	TTL      time.Duration // the time to live of the cached item
	Filepath string        // the filepath of the cached item
	err      chan error
}

func (c *item) Close() error {
	close(c.err)
	return nil
}

func newItem(key string, ttl time.Duration) (*item, error) {
	if ttl <= time.Second {
		return nil, fmt.Errorf("ttl '%s' is too short", ttl)
	}

	if err := IsValidKey(key); err != nil {
		return nil, err
	}

	var item = &item{
		Key:  key,
		Hash: strHash(key),
		TTL:  ttl,
		err:  make(chan error, 1),
	}

	return item, nil
}

func newItemKey(key string) (*item, error) {
	if err := IsValidKey(key); err != nil {
		return nil, err
	}

	var item = &item{
		Key:  key,
		Hash: strHash(key),
	}
	return item, nil
}

func IsValidKey(key string) error {
	if len(key) <= 1 {
		return fmt.Errorf("key '%s' is too short", key)
	}

	if !keyRegexFunc(key) {
		return fmt.Errorf("key '%s' contains invalid characters", key)
	}

	if len(key) > 64 {
		return fmt.Errorf("key '%s' is too long", key)
	}
	return nil
}

func (c *item) write(dir string, value []byte) {
	var (
		err      error
		path     string
		itemPath string
		file     *os.File
	)
	if c.TTL <= time.Second {
		c.err <- fmt.Errorf("ttl '%s' is too short", c.TTL)
	}

	path, itemPath = c.getpath(dir)
	if err = os.MkdirAll(path, 0755); err != nil {
		c.err <- err
		return
	}

	file, err = os.Create(itemPath)
	if err != nil {
		c.err <- err
		return
	}
	defer file.Close()

	_, err = file.Write(value)
	if err != nil {
		c.err <- err
		return
	}

	c.err <- nil
}

func (c *item) read(dir string) (value []byte, err error) {
	if c.TTL <= 0 {
		c.delete(dir)
		return nil, fmt.Errorf("item has expired: %d", c.TTL)
	}

	var _, itemPath = c.getpath(dir)

	var file *os.File
	file, err = os.Open(itemPath)
	if err != nil {
		return
	}
	defer file.Close()

	var stat os.FileInfo
	stat, err = file.Stat()
	if err != nil {
		return
	}

	value = make([]byte, stat.Size())
	_, err = file.Read(value)
	return
}

func (c *item) delete(dir string) (err error) {
	var path, itemPath = c.getpath(dir)
	err = os.Remove(itemPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return
	}

	return removeIfEmpty(path)
}

func removeIfEmpty(path string) (err error) {
	var files []fs.DirEntry
	files, err = os.ReadDir(path)
	if err != nil {
		return
	}

	if len(files) == 0 {
		err = os.Remove(path)
	}

	return
}

func (c *item) getpath(dir string) (path, itemPath string) {
	path = filepath.Join(dir, strconv.FormatUint(c.Hash, 10))
	itemPath = filepath.Join(path, c.Key)
	return path, itemPath
}

func (c *item) Equals(other *item) bool {
	return c.Key == other.Key
}

func (c *item) Gt(other *item) bool {
	return c.Key > other.Key
}

func (c *item) Lt(other *item) bool {
	return c.Key < other.Key
}

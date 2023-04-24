package src_test

import (
	"encoding/gob"
	"os"
	"testing"
	"time"

	"github.com/Nigel2392/netcache/src/client"
	"github.com/Nigel2392/netcache/src/logger"
	"github.com/Nigel2392/netcache/src/protocols"
	"github.com/Nigel2392/netcache/src/server"
)

var cacheServer = server.New("localhost", 8080, "./server-cache-test", time.Second*1, nil) // short timeout for testing (localhost)
var cacheClient = client.CacheClient{ServerAddr: "localhost:8080", Serializer: &protocols.XmlSerializer{}}

type testitem struct {
	Value   string `json:"value" xml:"value"` // interface{} for testing
	Keyable string `json:"keyable" xml:"keyable"`
}

func TestCacheServer(t *testing.T) {
	cacheServer.NewLogger(logger.Newlogger(logger.DEBUG, os.Stdout))
	go cacheServer.ListenAndServe()

	gob.Register(testitem{})

	var err = cacheClient.Connect()
	if err != nil {
		t.Fatal(err)
	}

	var items map[string]testitem = map[string]testitem{
		"key1": {Value: "value1", Keyable: "key1"},
		"key2": {Value: "value2", Keyable: "key2"},
		"key3": {Value: "value3", Keyable: "key3"},
		"key4": {Value: "value4", Keyable: "key4"},
		"key5": {Value: "value5", Keyable: "key5"},
	}

	for key, item := range items {
		err = cacheClient.Set(key, item, int64(5*time.Second))
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("LOG: set", key, item)
		}
	}

	for key, item := range items {
		var value testitem
		_, err = cacheClient.Get(key, &value)
		if err != nil {
			t.Fatal(err)
		}
		if value.Keyable != item.Keyable || value.Value != item.Value {
			t.Fatalf("value mismatch %s != %s or %s != %s", value.Keyable, item.Keyable, value.Value, item.Value)
		}
	}

	for key := range items {
		err = cacheClient.Delete(key)
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("LOG: deleted", key)
		}
	}

	for key := range items {
		var value testitem
		_, err = cacheClient.Get(key, &value)
		if err == nil {
			t.Fatalf("item not deleted %s", key)
		} else {
			t.Log("LOG: after delete, should error -", err)
		}
	}

	for key, item := range items {
		err = cacheClient.Set(key, item, int64(5*time.Second))
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("LOG: set", key, item)
		}
	}

	keys, err := cacheClient.Keys()
	if err != nil {
		t.Fatal(err)
	} else {
		for _, item := range items {
			var found bool
			for _, key := range keys {
				if key == item.Keyable {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("key not found %s", item.Keyable)
			}
		}
	}
	t.Log("LOG: waiting for items to expire")
	time.Sleep(6 * time.Second)
	t.Log("LOG: done waiting")

	for key := range items {
		var value testitem
		_, err = cacheClient.Get(key, &value)
		if err == nil {
			t.Fatalf("item not expired %s", key)
		} else {
			t.Log("LOG: after expiry, should error -", err)
		}
	}

	err = cacheClient.Clear()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("LOG: cleared")
	}

	cacheClient.Close()
}

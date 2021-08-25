package cache

import (
	"fmt"
	"testing"
)

func TestCache(t *testing.T) {
	var evicted int
	f := func(k, v interface{}) {
		evicted++
	}
	maxItems := 1500
	expectedEvicted := 500
	var cache Cache
	cache, err := New(Options{
		Size:      maxItems - expectedEvicted,
		OnEvicted: f,
	})
	if err != nil {
		t.Error("couldn't create the cache", err)
	}

	// set
	for i := 0; i < maxItems; i++ {
		key := fmt.Sprint(i)
		if err := cache.Set(key, key); err != nil {
			t.Errorf("item %s could not be set", fmt.Sprint(i))
		}
	}

	if evicted != expectedEvicted {
		t.Errorf("error during eviction: expected %d but got %d", expectedEvicted, evicted)
	}

	// get
	expectedRead := maxItems - expectedEvicted
	read := 0
	for i := 0; i < maxItems; i++ {
		key := fmt.Sprint(i)
		if v, ok := cache.Get(key); ok && key == fmt.Sprint(v) {
			read++
		}
	}
	if read != expectedRead {
		t.Errorf("error during eviction: expected %d but got %d", expectedEvicted, evicted)
	}

	// scan
	read = 0
	cache.Scan(func(_, _ interface{}) error {
		read++
		return nil
	})

	if read != expectedRead {
		t.Errorf("wrong number of items scanned: expected %d but got %d", read, expectedRead)
	}
}

package hybrid

import (
	"bytes"
	"fmt"
	"testing"
)

func TestHybrid(t *testing.T) {
	options := DefaultHybridOptions
	options.Cleanup = true
	evicted := 0
	options.OnEvicted = func(i1, i2 interface{}) {
		evicted++
	}
	hm, err := New(options)
	if err != nil {
		t.Error(err)
	}

	totalItems := 5000
	expectedEvicted := totalItems - options.MaxMemoryItem

	// write more items than the in memory max size
	for i := 0; i < totalItems; i++ {
		key := fmt.Sprint(i)
		if err := hm.Set(key, []byte(key)); err != nil {
			t.Error(err)
		}
	}

	// check the number of items moved to disk
	if evicted != expectedEvicted {
		t.Errorf("Total items evicted are different: expected %d but got %d", expectedEvicted, evicted)
	}

	// iterate all the items
	for i := 0; i < totalItems; i++ {
		key := fmt.Sprint(i)
		expectedValue := []byte(key)
		v, ok := hm.Get(key)
		if !ok {
			t.Errorf("item %s not found", key)
		} else if !bytes.EqualFold(expectedValue, v) {
			t.Errorf("item values are different: expected %s but got %s", expectedValue, v)
		}
	}
}

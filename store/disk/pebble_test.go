package disk

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestOpenPebbleDB(t *testing.T) {
	pbpath, _ := utiltestGetPath(t)
	pb, err := OpenPebbleDB(pbpath)
	if err != nil {
		t.Error(err)
	}
	utiltestRemoveDb(pb, pbpath)
}

func TestPebbleOperations(t *testing.T) {
	pbpath, _ := utiltestGetPath(t)
	pb, err := OpenPebbleDB(pbpath)
	if err != nil {
		t.Error(err)
	}

	maxItems := 100

	// set
	for i := 0; i < maxItems; i++ {
		key := fmt.Sprint(i)
		value := []byte(key)
		if err := pb.Set(key, value, time.Hour); err != nil {
			t.Error("[put] ", err)
		}
	}

	// get
	for i := 0; i < maxItems; i++ {
		key := fmt.Sprint(i)
		value := []byte(key)
		if data, err := pb.Get(key); err != nil || !bytes.EqualFold(data, value) {
			t.Errorf("[get] got %s but wanted %s: err %s", string(data), string(value), err)
		}
	}

	// scan
	read := 0
	pb.Scan(ScannerOptions{
		Handler: func(k, v []byte) error {
			_, _ = k, v
			read++
			return nil
		},
	})
	if read != maxItems {
		t.Errorf("[scan] got %d but wanted %d", read, maxItems)
	}

	// delete
	for i := 0; i < maxItems; i++ {
		key := fmt.Sprint(i)
		if err := pb.Del(key); err != nil {
			t.Errorf("[del] couldn't delete %s err %s", key, err)
		}
	}

	utiltestRemoveDb(pb, pbpath)
}

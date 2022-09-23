package disk

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"
)

type TestOperations struct {
	Set    bool
	Get    bool
	Scan   bool
	Delete bool
}

func utiltestOperations(t *testing.T, db DB, maxItems int, operations TestOperations) {
	// set
	if operations.Set {
		for i := 0; i < maxItems; i++ {
			key := fmt.Sprint(i)
			value := []byte(key)
			if err := db.Set(key, value, time.Hour); err != nil {
				t.Error("[put] ", err)
			}
		}
	}

	// get
	if operations.Get {
		for i := 0; i < maxItems; i++ {
			key := fmt.Sprint(i)
			value := []byte(key)
			if data, err := db.Get(key); err != nil || !bytes.EqualFold(data, value) {
				t.Errorf("[get] got %s but wanted %s: err %s", string(data), string(value), err)
			}
		}
	}

	// scan
	if operations.Scan {
		read := 0
		err := db.Scan(ScannerOptions{
			Handler: func(k, v []byte) error {
				_, _ = k, v
				read++
				return nil
			},
		})
		if err != nil {
			t.Error(err)
		}
		if read != maxItems {
			t.Errorf("[scan] got %d but wanted %d", read, maxItems)
		}
	}

	// delete
	if operations.Delete {
		for i := 0; i < maxItems; i++ {
			key := fmt.Sprint(i)
			if err := db.Del(key); err != nil {
				t.Errorf("[del] couldn't delete %s err %s", key, err)
			}
		}
	}
}

func utiltestGetPath(t *testing.T) (string, error) {
	tmpdir, err := os.MkdirTemp("", "hmaptest")
	if err != nil {
		t.Error(err)
	}
	return tmpdir, err
}

func utiltestRemoveDb(t *testing.T, db DB, pbpath string) {
	db.Close()
	os.RemoveAll(pbpath)
}

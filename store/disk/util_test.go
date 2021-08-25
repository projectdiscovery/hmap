package disk

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func utiltestOperations(t *testing.T, db DB, maxItems int) {
	// set
	for i := 0; i < maxItems; i++ {
		key := fmt.Sprint(i)
		value := []byte(key)
		if err := db.Set(key, value, time.Hour); err != nil {
			t.Error("[put] ", err)
		}
	}

	// get
	for i := 0; i < maxItems; i++ {
		key := fmt.Sprint(i)
		value := []byte(key)
		if data, err := db.Get(key); err != nil || !bytes.EqualFold(data, value) {
			t.Errorf("[get] got %s but wanted %s: err %s", string(data), string(value), err)
		}
	}

	// scan
	read := 0
	db.Scan(ScannerOptions{
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
		if err := db.Del(key); err != nil {
			t.Errorf("[del] couldn't delete %s err %s", key, err)
		}
	}
}

func utiltestGetPath(t *testing.T) (string, error) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "hmaptest")
	if err != nil {
		t.Error(err)
	}
	return tmpdir, err
}

func utiltestRemoveDb(t *testing.T, db DB, pbpath string) {
	db.Close()
	os.RemoveAll(pbpath)
}

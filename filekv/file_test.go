package filekv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/xid"
)

func TestFile(t *testing.T) {
	// add 200k items
	n := 100000
	var items1, items2 []string
	for i := 0; i < n; i++ {
		items1 = append(items1, xid.New().String())
		items2 = append(items2, xid.New().String())
	}

	dbFilename := filepath.Join(os.TempDir(), xid.New().String())
	options := DefaultOptions
	options.Path = dbFilename
	fdb, err := Open(options)
	if err != nil {
		t.Error(err)
	}
	defer fdb.Close()

	_, err = fdb.Merge(items1, items2)
	if err != nil {
		t.Error(err)
	}

	err = fdb.Process()
	if err != nil {
		t.Error(err)
	}

	allItems := append(items1, items2...)
	// all items should already exist
	for _, item := range allItems {
		err := fdb.Set([]byte(item), nil)
		if err == nil {
			t.Errorf("item %s doesn't exist\n", err)
		}
	}

	count := 0
	err = fdb.Scan(func(k, v []byte) error {
		// items should respect the input order
		ks := string(k)
		if !strings.EqualFold(ks, allItems[count]) {
			t.Errorf("item %s doesn't respect order\n", ks)
		}
		count++
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	expected := len(items1) + len(items2)
	if count != expected {
		t.Errorf("wrong number of items: wanted %d, got %d\n", expected, count)
	}
}

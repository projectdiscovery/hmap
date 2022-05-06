package disk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectdiscovery/fileutil"
	"github.com/projectdiscovery/hmap/filekv"
)

func TestKV(t *testing.T) {
	testOperations := TestOperations{
		Set:    true,
		Get:    true,
		Scan:   true,
		Delete: true,
	}
	var db DB
	// bbolt
	dbpath, _ := utiltestGetPath(t)
	dbpath = filepath.Join(dbpath, "boltdb")
	db, err := OpenBoltDBB(dbpath)
	if err != nil {
		t.Error(err)
	}
	db.(*BBoltDB).BucketName = "test"
	utiltestOperations(t, db, 100, testOperations)
	utiltestRemoveDb(t, db, dbpath)

	// pogreb
	dbpath, _ = utiltestGetPath(t)
	db, err = OpenPogrebDB(dbpath)
	if err != nil {
		t.Error(err)
	}
	utiltestOperations(t, db, 100, testOperations)
	utiltestRemoveDb(t, db, dbpath)

	// leveldb
	dbpath, _ = utiltestGetPath(t)
	db, err = OpenLevelDB(dbpath)
	if err != nil {
		t.Error(err)
	}
	utiltestOperations(t, db, 100, testOperations)
	utiltestRemoveDb(t, db, dbpath)
}

func TestFileKV(t *testing.T) {
	dbpath, _ := fileutil.GetTempFileName()
	os.RemoveAll(dbpath)
	opts := filekv.DefaultOptions
	opts.Cleanup = true
	opts.Compress = true
	opts.Path = dbpath

	fkv, err := filekv.Open(opts)
	if err != nil {
		t.Error(err)
	}
	_, _ = fkv.Merge([]string{"a", "b"}, []string{"b", "c"})
	_ = fkv.Process()
	count := 0
	_ = fkv.Scan(func(b1, b2 []byte) error {
		count++
		return nil
	})
	if count != 3 {
		t.Errorf("wanted 3 but got %d", count)
	}
}

package disk

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/projectdiscovery/hmap/filekv"
	fileutil "github.com/projectdiscovery/utils/file"
	stringsutil "github.com/projectdiscovery/utils/strings"
	"github.com/stretchr/testify/require"
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
	require.Nil(t, err)
	db.(*BBoltDB).BucketName = "test"
	utiltestOperations(t, db, 100, testOperations)
	utiltestRemoveDb(t, db, dbpath)

	// pogreb
	dbpath, _ = utiltestGetPath(t)
	db, err = OpenPogrebDB(dbpath)
	if runtime.GOOS == "windows" && stringsutil.EqualFoldAny(runtime.GOARCH, "arm", "arm64") {
		require.ErrorIs(t, ErrNotSupported, err)
	} else {
		require.Nil(t, err)
	}
	utiltestOperations(t, db, 100, testOperations)
	utiltestRemoveDb(t, db, dbpath)

	// leveldb
	dbpath, _ = utiltestGetPath(t)
	db, err = OpenLevelDB(dbpath)
	require.Nil(t, err)
	utiltestOperations(t, db, 100, testOperations)
	utiltestRemoveDb(t, db, dbpath)

	// buntdb
	dbpath, _ = utiltestGetPath(t)
	dbpath = filepath.Join(dbpath, "buntdb")
	db, err = OpenBuntDB(dbpath)
	require.Nil(t, err)
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
	require.Nil(t, err)
	_, _ = fkv.Merge([]string{"a", "b"}, []string{"b", "c"})
	_ = fkv.Process()
	count := 0
	_ = fkv.Scan(func(b1, b2 []byte) error {
		count++
		return nil
	})
	require.Equalf(t, 3, count, "wanted 3 but got %d", count)
}

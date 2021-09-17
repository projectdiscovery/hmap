package disk

import (
	"path/filepath"
	"testing"
)

func TestKV(t *testing.T) {
	var db DB
	// bbolt
	dbpath, _ := utiltestGetPath(t)
	dbpath = filepath.Join(dbpath, "boltdb")
	db, err := OpenBoltDBB(dbpath)
	if err != nil {
		t.Error(err)
	}
	db.(*BBoltDB).BucketName = "test"
	utiltestOperations(t, db, 100)
	utiltestRemoveDb(t, db, dbpath)

	// pebble
	dbpath, _ = utiltestGetPath(t)
	db, err = OpenPebbleDB(dbpath)
	if err != nil {
		t.Error(err)
	}
	utiltestOperations(t, db, 100)
	utiltestRemoveDb(t, db, dbpath)

	// pogreb
	dbpath, _ = utiltestGetPath(t)
	db, err = OpenPogrebDB(dbpath)
	if err != nil {
		t.Error(err)
	}
	utiltestOperations(t, db, 100)
	utiltestRemoveDb(t, db, dbpath)

	// leveldb
	dbpath, _ = utiltestGetPath(t)
	db, err = OpenLevelDB(dbpath)
	if err != nil {
		t.Error(err)
	}
	utiltestOperations(t, db, 100)
	utiltestRemoveDb(t, db, dbpath)

	// badgerdb
	dbpath, _ = utiltestGetPath(t)
	db, err = OpenBadgerDB(dbpath)
	if err != nil {
		t.Error(err)
	}
	utiltestOperations(t, db, 100)
	utiltestRemoveDb(t, db, dbpath)
}

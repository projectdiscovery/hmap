package disk

import (
	"bytes"
	"sync"
	"time"

	"github.com/projectdiscovery/hmap/filekv"
)

// FileDB - represents a file db implementation
type FileDB struct {
	db *filekv.FileDB
	sync.RWMutex
}

// OpenFileDB - Opens the specified path
func OpenFileDB(path string) (*FileDB, error) {
	options := filekv.DefaultOptions
	options.Path = path
	db, err := filekv.Open(options)
	if err != nil {
		return nil, err
	}

	fdb := new(FileDB)
	fdb.db = db

	return fdb, nil
}

// Size - returns the size of the database in bytes
func (fdb *FileDB) Size() int64 {
	return fdb.db.Size()
}

// Close ...
func (fdb *FileDB) Close() {
	fdb.db.Close()
}

// GC - runs the garbage collector
func (fdb *FileDB) GC() error {
	return ErrNotImplemented
}

// Incr - increment the key by the specified value
func (fdb *FileDB) Incr(k string, by int64) (int64, error) {
	return 0, ErrNotImplemented
}

func (fdb *FileDB) set(k, v []byte) error {
	return fdb.db.Set(k, v)
}

// Set - sets a key with the specified value and optional ttl
func (fdb *FileDB) Set(k string, v []byte, ttl time.Duration) error {
	return fdb.set([]byte(k), v)
}

// MSet - sets multiple key-value pairs
func (fdb *FileDB) MSet(data map[string][]byte) error {
	return ErrNotImplemented
}

// Get - fetches the value of the specified k
func (fdb *FileDB) Get(k string) ([]byte, error) {
	return nil, ErrNotImplemented
}

// MGet - fetch multiple values of the specified keys
func (fdb *FileDB) MGet(keys []string) [][]byte {
	return nil
}

// TTL - returns the time to live of the specified key's value
func (fdb *FileDB) TTL(key string) int64 {
	return 0
}

// MDel - removes key(s) from the store
func (fdb *FileDB) MDel(keys []string) error {
	return ErrNotImplemented
}

// Del - removes key from the store
func (fdb *FileDB) Del(key string) error {
	return ErrNotImplemented
}

// Scan - iterate over the whole store using the handler function
func (fdb *FileDB) Scan(scannerOpt ScannerOptions) error {
	valid := func(k []byte) bool {
		if k == nil {
			return false
		}

		if scannerOpt.Prefix != "" && !bytes.HasPrefix(k, []byte(scannerOpt.Prefix)) {
			return false
		}

		return true
	}

	return fdb.db.Scan(func(key, val []byte) error {
		if !valid(key) {
			return nil
		}
		return scannerOpt.Handler(key, val)
	})
}

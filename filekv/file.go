package filekv

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"io"
	"os"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/projectdiscovery/fileutil"
)

// FileDB - represents a file db implementation
type FileDB struct {
	stats       Stats
	options     Options
	tmpDbName   string
	tmpDb       *os.File
	tmpDbWriter io.WriteCloser
	db          *os.File
	dbWriter    io.WriteCloser
	mdb         *lru.Cache
	sync.RWMutex
}

// Open a new file based db
func Open(options Options) (*FileDB, error) {
	db, err := os.OpenFile(options.Path, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}

	tmpFileName, err := fileutil.GetTempFileName()
	if err != nil {
		return nil, err
	}
	tmpDb, err := os.Create(tmpFileName)
	if err != nil {
		return nil, err
	}

	fdb := &FileDB{
		tmpDbName: tmpFileName,
		options:   options,
		db:        db,
		tmpDb:     tmpDb,
	}

	if options.Dedupe {
		fdb.mdb, err = lru.New(int(options.MaxItems))
		if err != nil {
			return nil, err
		}
	}

	if options.Compress {
		fdb.tmpDbWriter = zlib.NewWriter(fdb.tmpDb)
		fdb.dbWriter = zlib.NewWriter(fdb.db)
	} else {
		fdb.tmpDbWriter = fdb.tmpDb
		fdb.dbWriter = fdb.db
	}

	return fdb, nil
}

// Process added files/slices/elements
func (fdb *FileDB) Process() error {
	// Closes the temporary file
	if fdb.options.Compress {
		// close the writer
		if err := fdb.tmpDbWriter.Close(); err != nil {
			return err
		}
	}

	// closes the file to flush to disk and reopen it
	_ = fdb.tmpDb.Close()
	var err error
	fdb.tmpDb, err = os.Open(fdb.tmpDbName)
	if err != nil {
		return err
	}

	var tmpDbReader io.Reader
	if fdb.options.Compress {
		var err error
		tmpDbReader, err = zlib.NewReader(fdb.tmpDb)
		if err != nil {
			return err
		}
	} else {
		tmpDbReader = fdb.tmpDb
	}

	sc := bufio.NewScanner(tmpDbReader)
	buf := make([]byte, BufferSize)
	sc.Buffer(buf, BufferSize)
	for sc.Scan() {
		_ = fdb.Set(sc.Bytes(), nil)
	}

	fdb.tmpDb.Close()

	// flush to disk
	fdb.dbWriter.Close()
	fdb.db.Close()

	return nil
}

// Reset the db
func (fdb *FileDB) Reset() error {
	// clear the cache
	if fdb.options.Dedupe {
		fdb.mdb.Purge()
	}

	// reset the tmp file
	fdb.tmpDb.Close()
	var err error
	fdb.tmpDb, err = os.Create(fdb.tmpDbName)
	if err != nil {
		return err
	}

	// reset the target file
	fdb.db.Close()
	fdb.db, err = os.Create(fdb.tmpDbName)
	if err != nil {
		return err
	}

	if fdb.options.Compress {
		fdb.tmpDbWriter = zlib.NewWriter(fdb.tmpDb)
		fdb.dbWriter = zlib.NewWriter(fdb.db)
	} else {
		fdb.tmpDbWriter = fdb.tmpDb
		fdb.dbWriter = fdb.db
	}

	return nil
}

// Size - returns the size of the database in bytes
func (fdb *FileDB) Size() int64 {
	osstat, err := fdb.db.Stat()
	if err != nil {
		return 0
	}
	return osstat.Size()
}

// Close ...
func (fdb *FileDB) Close() {
	tmpDBFilename := fdb.tmpDb.Name()
	_ = fdb.tmpDb.Close()
	os.RemoveAll(tmpDBFilename)

	_ = fdb.db.Close()
	dbFilename := fdb.db.Name()
	if fdb.options.Cleanup {
		os.RemoveAll(dbFilename)
	}
}

func (fdb *FileDB) set(k, v []byte) error {
	var s bytes.Buffer
	s.Write(k)
	s.WriteString(Separator)
	s.Write(v)
	s.WriteString(NewLine)
	_, err := fdb.dbWriter.Write(s.Bytes())
	if err != nil {
		return err
	}
	fdb.stats.NumberOfItems++
	return nil
}

func (fdb *FileDB) Set(k, v []byte) error {
	// check for duplicates
	if fdb.options.Dedupe {
		if ok, _ := fdb.mdb.ContainsOrAdd(string(k), struct{}{}); ok {
			fdb.stats.NumberOfDupedItems++
			return ErrItemExists
		}
	}

	if fdb.shouldSkip(k, v) {
		fdb.stats.NumberOfFilteredItems++
		return ErrItemFiltered
	}

	fdb.stats.NumberOfItems++
	return fdb.set(k, v)
}

// Scan - iterate over the whole store using the handler function
func (fdb *FileDB) Scan(handler func([]byte, []byte) error) error {
	// open the db and scan
	dbCopy, err := os.Open(fdb.options.Path)
	if err != nil {
		return err
	}
	defer dbCopy.Close()

	var dbReader io.ReadCloser
	if fdb.options.Compress {
		dbReader, err = zlib.NewReader(dbCopy)
		if err != nil {
			return err
		}
	} else {
		dbReader = dbCopy
	}

	sc := bufio.NewScanner(dbReader)
	buf := make([]byte, BufferSize)
	sc.Buffer(buf, BufferSize)
	for sc.Scan() {
		tokens := bytes.SplitN(sc.Bytes(), []byte(Separator), 2)
		var k, v []byte
		if len(tokens) > 0 {
			k = tokens[0]
		}
		if len(tokens) > 1 {
			v = tokens[1]
		}
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

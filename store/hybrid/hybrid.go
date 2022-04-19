package hybrid

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/projectdiscovery/fileutil"
	"github.com/projectdiscovery/hmap/store/cache"
	"github.com/projectdiscovery/hmap/store/disk"
	"github.com/projectdiscovery/stringsutil"
)

type MapType int

const (
	Memory MapType = iota
	Disk
	Hybrid
)

type DBType int

const (
	LevelDB DBType = iota
	PogrebDB
	BBoltDB
	// FileDB preserves input order
	FileDB
)

type Options struct {
	MemoryExpirationTime   time.Duration
	DiskExpirationTime     time.Duration
	Type                   MapType
	DBType                 DBType
	MoveToDiskOnExpiration bool
	Path                   string
	Cleanup                bool
	MaxMemoryItem          int
	OnEvicted              func(interface{}, interface{})
	// Remove temporary hmap in the temporary folder older than duration
	RemoveOlderThan time.Duration
}

var DefaultMemoryOptions = Options{
	Type:          Memory,
	MaxMemoryItem: 2500,
}

var DefaultDiskOptions = Options{
	Type:            Disk,
	DBType:          LevelDB,
	Cleanup:         true,
	RemoveOlderThan: 24 * time.Hour * 2, // at startup removes temporary dbs older than x days
}

var DefaultHybridOptions = Options{
	Type:          Hybrid,
	DBType:        PogrebDB,
	MaxMemoryItem: 2500,
}

type HybridMap struct {
	options     *Options
	memorymap   cache.Cache
	diskmap     disk.DB
	diskmapPath string
}

func New(options Options) (*HybridMap, error) {
	executableName := fileutil.ExecutableName()

	// Due to potential system failures, the first operation is removing leftovers older than the defined duration
	// if cleanup is true and a max age duration has been defined
	if options.Cleanup && options.Path == "" && options.RemoveOlderThan > 0 {
		targetCleanupDir := os.TempDir()
		tmpFiles, err := os.ReadDir(targetCleanupDir)
		if err != nil {
			return nil, err
		}
		now := time.Now()
		for _, tmpFile := range tmpFiles {
			// discard non folders
			if !tmpFile.IsDir() {
				continue
			}
			// discard folders not containing the executable name
			if !stringsutil.ContainsAny(tmpFile.Name(), executableName) {
				continue
			}

			tmpFileInfo, err := tmpFile.Info()
			if err != nil {
				continue
			}
			modTime := tmpFileInfo.ModTime()
			if now.Sub(modTime) > options.RemoveOlderThan {
				targetFolderFullPath := filepath.Join(targetCleanupDir, tmpFileInfo.Name())
				os.RemoveAll(targetFolderFullPath)
			}
		}
	}

	var hm HybridMap
	if options.Type == Memory || options.Type == Hybrid {
		var err error
		hm.memorymap, err = cache.New(cache.Options{
			Duration: options.MemoryExpirationTime,
			Size:     options.MaxMemoryItem,
			OnEvicted: func(k, v interface{}) {
				if options.Type == Hybrid {
					hm.diskmap.Set(fmt.Sprint(k), v.([]byte), 0)
				}
				if options.OnEvicted != nil {
					options.OnEvicted(k, v)
				}
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if options.Type == Disk || options.Type == Hybrid {
		diskmapPathm := options.Path
		if diskmapPathm == "" {
			var err error
			diskmapPathm, err = ioutil.TempDir("", executableName)
			if err != nil {
				return nil, err
			}
		}

		hm.diskmapPath = diskmapPathm
		switch options.DBType {
		case PogrebDB:
			db, err := disk.OpenPogrebDB(diskmapPathm)
			if err != nil {
				return nil, err
			}
			hm.diskmap = db
		case BBoltDB:
			db, err := disk.OpenBoltDBB(filepath.Join(diskmapPathm, "bb"))
			if err != nil {
				return nil, err
			}
			hm.diskmap = db
		case FileDB:
			db, err := disk.OpenFileDB(filepath.Join(diskmapPathm, "ff"))
			if err != nil {
				return nil, err
			}
			hm.diskmap = db
		case LevelDB:
			fallthrough
		default:
			db, err := disk.OpenLevelDB(diskmapPathm)
			if err != nil {
				return nil, err
			}
			hm.diskmap = db
		}
	}

	hm.options = &options

	return &hm, nil
}

func (hm *HybridMap) Close() error {
	if hm.diskmap != nil {
		hm.diskmap.Close()
	}
	if hm.diskmapPath != "" && hm.options.Cleanup {
		return os.RemoveAll(hm.diskmapPath)
	}
	return nil
}

func (hm *HybridMap) Set(k string, v []byte) error {
	var err error
	switch hm.options.Type {
	case Hybrid:
		fallthrough
	case Memory:
		hm.memorymap.Set(k, v)
	case Disk:
		err = hm.diskmap.Set(k, v, hm.options.DiskExpirationTime)
	}

	return err
}

func (hm *HybridMap) Get(k string) ([]byte, bool) {
	switch hm.options.Type {
	case Memory:
		v, ok := hm.memorymap.Get(k)
		if ok {
			return v.([]byte), ok
		}
		return []byte{}, ok
	case Hybrid:
		v, ok := hm.memorymap.Get(k)
		if ok {
			return v.([]byte), ok
		}
		vm, err := hm.diskmap.Get(k)
		// load it in memory since it has been recently used
		if err == nil {
			hm.memorymap.Set(k, vm)
		}
		return vm, err == nil
	case Disk:
		v, err := hm.diskmap.Get(k)
		return v, err == nil
	}

	return []byte{}, false
}

func (hm *HybridMap) Del(key string) error {
	switch hm.options.Type {
	case Memory:
		hm.memorymap.Delete(key)
	case Hybrid:
		hm.memorymap.Delete(key)
		return hm.diskmap.Del(key)
	case Disk:
		return hm.diskmap.Del(key)
	}

	return nil
}

func (hm *HybridMap) Scan(f func(interface{}, interface{}) error) {
	switch hm.options.Type {
	case Memory:
		hm.memorymap.Scan(f)
	case Hybrid:
		hm.memorymap.Scan(f)
		_ = hm.diskmap.Scan(disk.ScannerOptions{Handler: f})
	case Disk:
		_ = hm.diskmap.Scan(disk.ScannerOptions{Handler: f})
	}
}

func (hm *HybridMap) Size() int64 {
	var count int64
	if hm.memorymap != nil {
		count += int64(hm.memorymap.ItemCount())
	}
	if hm.diskmap != nil {
		count += hm.diskmap.Size()
	}
	return count
}

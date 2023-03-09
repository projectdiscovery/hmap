# hmap

Hybrid memory/disk map that helps you to manage key value storage for input deduplication.

---

Available functions:

|Name|Declaration/Params/Return|
|-|-|
|New|func New(options Options) (*HybridMap, error){}|
|Close|func (hm *HybridMap) Close() error{}|
|Set|func (hm *HybridMap) Set(k string, v []byte) error{}|
|Get|func (hm *HybridMap) Get(k string) ([]byte, bool){}|
|Del|func (hm *HybridMap) Del(key string) error{}|
|Scan|func (hm *HybridMap) Scan(f func([]byte, []byte) error){}|
|Size|func (hm *HybridMap) Size() int64{}|
|TuneMemory|func (hm *HybridMap) TuneMemory(){}|

Available options:

```go
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
	BuntDB
)
```

|Name|Props|
|-|-|
|`DefaultOptions`|- Type: Memory<br>- MemoryExpirationTime: time.Duration(5) * time.Minute<br>- JanitorTime:          time.Duration(1) * time.Minute|
|`DefaultMemoryOptions`|- Type: Memory|
|`DefaultDiskOptions`|- Type: Disk<br>- DBType: LevelDB<br>- Cleanup: true<br>- RemoveOlderThan: 24* time.Hour *2|
|`DefaultDiskOptions`|- Type: Hybrid<br>- DBType: PogrebDB<br>- MemoryExpirationTime: time.Duration(5) * time.Minute<br>- JanitorTime: time.Duration(1) * time.Minute|

Custom options:
```go
type Options struct {
	MemoryExpirationTime time.Duration
	DiskExpirationTime   time.Duration
	JanitorTime          time.Duration
	Type                 MapType
	DBType               DBType
	MemoryGuardForceDisk bool
	MemoryGuard          bool
	MaxMemorySize        int
	MemoryGuardTime      time.Duration
	Path                 string
	Cleanup              bool
	Name                 string
	RemoveOlderThan time.Duration
}
```

# Simple usage example

```go
func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go normal(&wg)
	wg.Wait()
}

func normal(wg *sync.WaitGroup) {
	defer wg.Done()
	hm, err := hybrid.New(hybrid.DefaultOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer hm.Close()
	err2 := hm.Set("a", []byte("b"))
	if err2 != nil {
		log.Fatal(err2)
	}
	v, ok := hm.Get("a")
	if ok {
		log.Println(v)
	}
}
```

# License
hmap is distributed under MIT License
// this is a wrapper around gcache - the eviction strategy is used to move data to disk
package cache

import (
	"fmt"
	"math"
	"time"

	"github.com/bluele/gcache"
)

type Options struct {
	Duration            time.Duration
	Size                int
	LRU                 bool
	LFU                 bool
	AdaptiveReplacement bool
	OnEvicted           func(interface{}, interface{})
}

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type Cache interface {
	SetWithExpiration(string, interface{}, time.Duration)
	Set(string, interface{})
	Get(string) (interface{}, bool)
	Delete(string) (interface{}, bool)
	CloneItems() map[string]Item
	Scan(func([]byte, []byte) error)
	ItemCount() int
}

type CacheMemory struct {
	DefaultExpiration time.Duration
	cache             gcache.Cache
	Items             map[string]Item
	onEvicted         func(string, interface{})
	janitor           *janitor
}

func (c *CacheMemory) SetWithExpiration(k string, x interface{}, d time.Duration) {
	return c.set(k, x, d)
}

func (c *CacheMemory) set(k string, x interface{}, d time.Duration) error {
	if d <= 0 {
		c.cache.Set(k, x)
	}
	return c.cache.SetWithExpire(k, x, d)
}

func (c *CacheMemory) Set(k string, x interface{}) error {
	return c.set(k, x, 0)
}

func (c *CacheMemory) Get(k string) (interface{}, bool) {
	if !c.cache.Has(k) {
		return nil, false
	}
	value, err := c.cache.Get(k)
	if err != nil {
		return value, false
	}
	return value, true
}

func (c *CacheMemory) get(k string) (interface{}, bool) {
	item, found := c.Items[k]
	if !found {
		return nil, false
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}

	return item.Object, true
}

func (c *CacheMemory) Delete(k string) (interface{}, bool) {
	return c.delete(k)
}

func (c *CacheMemory) delete(k string) (interface{}, bool) {
	v, err := c.cache.GetIFPresent(k)
	if err != nil {
		return nil, false
	}
	if c.cache.Remove(k) {
		c.onEvicted(k, v)
		return v, true
	}

	return nil, false
}

func (c *CacheMemory) OnEvicted(f func(string, interface{})) {
	c.onEvicted = f
}

func (c *CacheMemory) Scan(f func(interface{}, interface{}) error) {
	for k, v := range c.cache.GetALL(true) {
		f(k, v)
	}
}

func (c *CacheMemory) CloneItems() map[string]interface{} {
	items := make(map[string]interface{})
	for k, v := range c.cache.GetALL(true) {
		items[fmt.Sprint(k)] = v
	}
	return items
}

func (c *CacheMemory) ItemCount() int {
	return c.cache.Len(true)
}

func (c *CacheMemory) Empty() {
	c.cache.Purge()
}

func New(options Options) *CacheMemory {
	return NewFrom(options, nil)
}

func NewFrom(options Options, items map[string]Item) *CacheMemory {
	maxSize := math.MaxInt32
	if options.Size > 0 {
		maxSize = options.Size
	}
	cache := gcache.New(maxSize)
	if options.Duration > 0 {
		cache = cache.Expiration(options.Duration)
	}
	if options.AdaptiveReplacement {
		cache = cache.ARC()
	} else if options.LRU {
		cache = cache.LRU()
	} else if options.LFU {
		cache = cache.LFU()
	}
	return &CacheMemory{cache: cache.Build()}
}

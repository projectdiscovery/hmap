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
	SetWithExpiration(string, interface{}, time.Duration) error
	Set(string, interface{}) error
	Get(string) (interface{}, bool)
	Delete(string) (interface{}, bool)
	CloneItems() map[string]interface{}
	Scan(func(interface{}, interface{}) error)
	ItemCount() int
}

type CacheMemory struct {
	DefaultExpiration time.Duration
	cache             gcache.Cache
	onEvicted         func(string, interface{})
}

func (c *CacheMemory) SetWithExpiration(k string, x interface{}, d time.Duration) error {
	return c.set(k, x, d)
}

func (c *CacheMemory) set(k string, x interface{}, d time.Duration) error {
	if d <= 0 {
		return c.cache.Set(k, x)
	}
	return c.cache.SetWithExpire(k, x, d)
}

func (c *CacheMemory) Set(k string, x interface{}) error {
	return c.set(k, x, 0)
}

func (c *CacheMemory) Get(k string) (interface{}, bool) {
	return c.get(k)
}

func (c *CacheMemory) get(k string) (interface{}, bool) {
	if !c.cache.Has(k) {
		return nil, false
	}
	value, err := c.cache.Get(k)
	if err != nil {
		return value, false
	}
	return value, true
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

func New(options Options) (*CacheMemory, error) {
	return NewFrom(options, nil)
}

func NewFrom(options Options, items map[string]interface{}) (*CacheMemory, error) {
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

	if options.OnEvicted != nil {
		cache.EvictedFunc(options.OnEvicted)
	}

	builtCache := cache.Build()
	for k, v := range items {
		if err := builtCache.Set(k, v); err != nil {
			return nil, err
		}
	}
	return &CacheMemory{cache: builtCache}, nil
}

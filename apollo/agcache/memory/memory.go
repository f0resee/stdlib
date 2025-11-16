package memory

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/f0resee/stdlib/apollo/agcache"
)

type DefaultCache struct {
	defaultCache sync.Map
	count        int64
}

func (d *DefaultCache) Set(key string, value interface{}, expireSeconds int) (err error) {
	d.defaultCache.Store(key, value)
	atomic.AddInt64(&d.count, int64(1))
	return nil
}

func (d *DefaultCache) EntryCount() (entryCount int64) {
	c := atomic.LoadInt64(&d.count)
	return c
}

func (d *DefaultCache) Get(key string) (value interface{}, err error) {
	v, ok := d.defaultCache.Load(key)
	if !ok {
		return nil, errors.New("load default cache fail")
	}
	return v, nil
}

func (d *DefaultCache) Range(f func(key, value interface{}) bool) {
	d.defaultCache.Range(f)
}

func (d *DefaultCache) Del(key string) (affected bool) {
	d.defaultCache.Delete(key)
	atomic.AddInt64(&d.count, int64(-1))
	return true
}

func (d *DefaultCache) Clear() {
	d.defaultCache = sync.Map{}
	atomic.StoreInt64(&d.count, int64(0))
}

type DefaultCacheFactory struct {
}

func (d *DefaultCacheFactory) Create() agcache.CacheInterface {
	return &DefaultCache{}
}

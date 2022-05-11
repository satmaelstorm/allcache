package allcache

import (
	"github.com/satmaelstorm/list"
	"sync"
)

type LFU[K comparable, T any] struct {
	cache *ntsLFU[K, T]
	lock  sync.Mutex
}

func NewLFU[K comparable, T any](maxSize int) Cache[K, T] {
	cache := new(LFU[K, T])
	cache.cache = newNtsLFU[K, T](maxSize)
	return cache
}

func (c *LFU[K, T]) Put(key K, item T) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.put(key, item)
}

func (c *LFU[K, T]) Get(key K, def T) (T, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.cache.get(key, def)
}

func (c *LFU[K, T]) Delete(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.delete(key)
}

type ntsLFU[K comparable, T any] struct {
	items      map[K]*list.PqItem[int64, cacheEntry[K, T]]
	evictQueue *list.PQ[int64, cacheEntry[K, T]]

	maxSize int
}

func newNtsLFU[K comparable, T any](maxSize int) *ntsLFU[K, T] {
	return &ntsLFU[K, T]{
		items:      make(map[K]*list.PqItem[int64, cacheEntry[K, T]], maxSize),
		evictQueue: list.NewPQ[int64, cacheEntry[K, T]](maxSize),
	}
}

func (c *ntsLFU[K, T]) put(key K, value T) {
	if e, ok := c.items[key]; ok {
		entry := e.GetValue()
		entry.value = value
		e.SetValue(entry)
		return
	}
	added, oust := c.evictQueue.EnqueueWithOust(-1, cacheEntry[K, T]{key: key, value: value})
	if oust != nil {
		delete(c.items, oust.GetValue().key)
	}
	c.items[key] = added
}

func (c *ntsLFU[K, T]) get(key K, def T) (T, bool) {
	if e, ok := c.items[key]; ok {
		c.evictQueue.DecInPosition(e.GetIndex())
		return e.GetValue().value, true
	}
	return def, false
}

func (c *ntsLFU[K, T]) delete(key K) {
	if e, ok := c.items[key]; ok {
		delete(c.items, key)
		c.evictQueue.Delete(e)
	}
}

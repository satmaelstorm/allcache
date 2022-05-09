package allcache

import (
	"github.com/satmaelstorm/list"
	"sync"
)

type LRU[K comparable, T any] struct {
	lru  *ntsLRU[K, T]
	lock sync.Mutex
}

func NewLRU[K comparable, T any](
	maxSize int64,
	calcSize SizeCalculator[T],
) Cache[K, T] {
	lru := new(LRU[K, T])
	lru.lru = newNtsLRU[K, T](maxSize, calcSize)
	return lru
}

func (c *LRU[K, T]) Put(key K, item T) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.lru.put(key, item)
}

func (c *LRU[K, T]) Get(key K, def T) (T, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.get(key, def)
}

func (c *LRU[K, T]) Delete(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.lru.delete(key)
}

//non thread safe LRU
type ntsLRU[K comparable, T any] struct {
	items      map[K]*list.Node[cacheEntry[K, T]]
	evictQueue *list.Queue[cacheEntry[K, T]]
	length     int64
	maxSize    int64
	sizeCalc   SizeCalculator[T]
}

func newNtsLRU[K comparable, T any](
	maxSize int64,
	sizeCalc SizeCalculator[T],
) *ntsLRU[K, T] {
	if nil == sizeCalc {
		sizeCalc = func(T) int64 { return 1 }
	}
	return &ntsLRU[K, T]{
		items:      make(map[K]*list.Node[cacheEntry[K, T]], maxSize),
		evictQueue: list.NewQueue[cacheEntry[K, T]](),
		maxSize:    maxSize,
		length:     0,
		sizeCalc:   sizeCalc,
	}
}

func (c *ntsLRU[K, T]) put(key K, value T) {
	defer c.adjust()
	newValue := cacheEntry[K, T]{key, value}
	if e, ok := c.items[key]; ok {
		c.evictQueue.MoveToBack(e)
		oldValue := e.Value().value
		e.SetValue(newValue)
		c.length += c.sizeCalc(value) - c.sizeCalc(oldValue)
		return
	}
	c.evictQueue.Enqueue(newValue)
	c.length += c.sizeCalc(value)
	c.items[key] = c.evictQueue.Tail()
	return
}

func (c *ntsLRU[K, T]) evict() {
	e := c.evictQueue.Dequeue()
	if e != nil {
		c.remove(e)
	}
}

func (c *ntsLRU[K, T]) remove(e *list.Node[cacheEntry[K, T]]) {
	delete(c.items, e.Value().key)
	c.length -= c.sizeCalc(e.Value().value)
}

func (c *ntsLRU[K, T]) adjust() {
	for {
		if c.maxSize >= c.length {
			return
		}
		c.evict()
	}
}

func (c *ntsLRU[K, T]) get(key K, def T) (T, bool) {
	if e, ok := c.items[key]; ok {
		c.evictQueue.MoveToBack(e)
		return e.Value().value, ok
	}
	return def, false
}

func (c *ntsLRU[K, T]) delete(key K) {
	if e, ok := c.items[key]; ok {
		c.remove(e)
	}
}

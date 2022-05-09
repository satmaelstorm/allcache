package allcache

import (
	"github.com/satmaelstorm/list"
	"sync"
)

// Simplified2Q - simplified2Q @see http://www.vldb.org/conf/1994/P439.PDF
type Simplified2Q[K comparable, T any] struct {
	cache *ntsSimplified2Q[K, T]
	lock  sync.Mutex
}

func NewSimplified2Q[K comparable, T any](a1Size, amSize int64) Cache[K, T] {
	cache := new(Simplified2Q[K, T])
	cache.cache = newNtsSimplified2Q[K, T](a1Size, amSize)
	return cache
}

func (c *Simplified2Q[K, T]) Put(key K, item T) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.put(key, item)
}

func (c *Simplified2Q[K, T]) Get(key K, def T) (T, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.cache.get(key, def)
}

func (c *Simplified2Q[K, T]) Delete(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.delete(key)
}

//non thread safe Simplified 2Q
//@see http://www.vldb.org/conf/1994/P439.PDF
type ntsSimplified2Q[K comparable, T any] struct {
	items     map[K]*list.Node[cacheEntry2Q[K, T]]
	am        *list.Queue[cacheEntry2Q[K, T]]
	a1        *list.Queue[cacheEntry2Q[K, T]]
	a1Size    int64
	amSize    int64
	a1Length  int64
	amLength  int64
	totalSize int64
}

func newNtsSimplified2Q[K comparable, T any](a1Size, amSize int64) *ntsSimplified2Q[K, T] {
	return &ntsSimplified2Q[K, T]{
		items:     make(map[K]*list.Node[cacheEntry2Q[K, T]], a1Size+amSize),
		am:        list.NewQueue[cacheEntry2Q[K, T]](),
		a1:        list.NewQueue[cacheEntry2Q[K, T]](),
		a1Size:    a1Size,
		amSize:    amSize,
		totalSize: a1Size + amSize,
	}
}

func (c *ntsSimplified2Q[K, T]) get(key K, def T) (T, bool) {
	if e, ok := c.items[key]; ok {
		if e.Value().isAm {
			c.am.MoveToBack(e)
		} else {
			c.a1.Remove(e)
			cacheEntry := e.Value()
			cacheEntry.isAm = true
			e.SetValue(cacheEntry)
			c.am.Enqueue(cacheEntry)
			c.items[key] = c.am.Tail()
		}
		return e.Value().value, true
	}
	return def, false
}

func (c *ntsSimplified2Q[K, T]) put(key K, value T) {
	if e, ok := c.items[key]; ok {
		cacheEntry := e.Value()
		cacheEntry.value = value
		if cacheEntry.isAm {
			e.SetValue(cacheEntry)
			c.am.MoveToBack(e)
		} else {
			c.a1.Remove(e)
			cacheEntry.isAm = true
			e.SetValue(cacheEntry)
			c.am.Enqueue(cacheEntry)
			c.items[key] = c.am.Tail()
		}
		return
	}

	cacheEntry := cacheEntry2Q[K, T]{isAm: false}
	cacheEntry.key = key
	cacheEntry.value = value

	defer func() {
		c.a1.Enqueue(cacheEntry)
		c.items[key] = c.a1.Tail()
	}()

	if c.totalSize > int64(c.a1.Len()+c.am.Len()) {
		return
	}

	if c.a1Size <= int64(c.a1.Len()) {
		e := c.a1.Dequeue()
		if e != nil {
			delete(c.items, e.Value().key)
		}
		return
	}

	e := c.am.Dequeue()
	if e != nil {
		delete(c.items, e.Value().key)
	}
}

func (c *ntsSimplified2Q[K, T]) delete(key K) {
	if e, ok := c.items[key]; ok {
		c.remove(e)
	}
}

func (c *ntsSimplified2Q[K, T]) remove(e *list.Node[cacheEntry2Q[K, T]]) {
	delete(c.items, e.Value().key)
	if e.Value().isAm {
		c.am.Remove(e)
	} else {
		c.a1.Remove(e)
	}
}

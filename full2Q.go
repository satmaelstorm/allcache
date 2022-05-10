package allcache

import (
	"github.com/satmaelstorm/list"
	"sync"
)

// Full2Q - full version 2Q - @see http://www.vldb.org/conf/1994/P439.PDF
type Full2Q[K comparable, T any] struct {
	cache *ntsFull2Q[K, T]
	lock  sync.Mutex
}

func NewFull2Q[K comparable, T any](amSize, a1InSize, a1OutSize uint64) Cache[K, T] {
	cache := new(Full2Q[K, T])
	cache.cache = newNtsFull2Q[K, T](amSize, a1InSize, a1OutSize)
	return cache
}

func (c *Full2Q[K, T]) Put(key K, item T) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.put(key, item)
}

func (c *Full2Q[K, T]) Get(key K, def T) (T, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.cache.get(key, def)
}

func (c *Full2Q[K, T]) Delete(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.delete(key)
}

//non thead safe full version 2Q - @see http://www.vldb.org/conf/1994/P439.PDF
type ntsFull2Q[K comparable, T any] struct {
	items map[K]*list.Node[cacheEntry2Q[K, T]]
	am    *list.Queue[cacheEntry2Q[K, T]]
	a1in  *list.Queue[cacheEntry2Q[K, T]]

	itemsOut map[K]*list.Node[K]
	a1out    *list.Queue[K]

	amSize    uint64
	a1InSize  uint64
	a1OutSize uint64
	totalSize uint64
}

func newNtsFull2Q[K comparable, T any](amSize, a1InSize, a1OutSize uint64) *ntsFull2Q[K, T] {
	return &ntsFull2Q[K, T]{
		items: make(map[K]*list.Node[cacheEntry2Q[K, T]], amSize+a1InSize),
		am:    list.NewQueue[cacheEntry2Q[K, T]](),
		a1in:  list.NewQueue[cacheEntry2Q[K, T]](),

		itemsOut: make(map[K]*list.Node[K], a1OutSize),
		a1out:    list.NewQueue[K](),

		amSize:    amSize,
		a1InSize:  a1InSize,
		a1OutSize: a1OutSize,

		totalSize: amSize + a1InSize,
	}
}

func (c *ntsFull2Q[K, T]) get(key K, def T) (T, bool) {
	if e, ok := c.items[key]; ok {
		if e.Value().isAm {
			c.am.MoveToBack(e)
		}
		return e.Value().value, true
	}
	return def, false
}

func (c *ntsFull2Q[K, T]) put(key K, value T) {
	if e, ok := c.items[key]; ok {
		cacheEntry := e.Value()
		cacheEntry.value = value
		if cacheEntry.isAm {
			e.SetValue(cacheEntry)
			c.am.MoveToBack(e)
		} else {
			e.SetValue(cacheEntry)
		}
		return
	}

	c.reclaim()

	cacheEntry := cacheEntry2Q[K, T]{isAm: false}
	cacheEntry.key = key
	cacheEntry.value = value

	if _, ok := c.itemsOut[key]; ok {
		cacheEntry.isAm = true
	}

	if cacheEntry.isAm {
		c.am.Enqueue(cacheEntry)
		c.items[cacheEntry.key] = c.am.Tail()
	} else {
		c.a1in.Enqueue(cacheEntry)
		c.items[cacheEntry.key] = c.a1in.Tail()
	}
}

func (c *ntsFull2Q[K, T]) reclaim() {
	if c.totalSize > uint64(c.am.Len()+c.a1in.Len()) {
		return //there are free slots
	}
	if uint64(c.a1in.Len()) > c.a1InSize {
		y := c.a1in.Dequeue()
		if nil == y {
			return
		}
		delete(c.items, y.Value().key)
		c.a1out.Enqueue(y.Value().key)
		c.itemsOut[y.Value().key] = c.a1out.Tail()
		if uint64(c.a1out.Len()) > c.a1OutSize {
			z := c.a1out.Dequeue()
			if z != nil {
				delete(c.itemsOut, z.Value())
			}
		}
	} else {
		y := c.am.Dequeue()
		if y != nil {
			delete(c.items, y.Value().key)
		}
	}
}

func (c *ntsFull2Q[K, T]) delete(key K) {
	if e, ok := c.items[key]; ok {
		if e.Value().isAm {
			c.am.Remove(e)
		} else {
			c.a1in.Remove(e)
		}
		delete(c.items, key)
	}
	if e, ok := c.itemsOut[key]; ok {
		c.a1out.Remove(e)
		delete(c.itemsOut, key)
	}
}

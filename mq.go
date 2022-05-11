package allcache

import (
	"github.com/satmaelstorm/list"
	"math"
	"sync"
)

type MQ[K comparable, T any] struct {
	cache *ntsMqCache[K, T]
	lock  sync.Mutex
}

func NewMQCache[K comparable, T any](
	queues byte,
	maxSize, qOutSize, lifeTime uint64,
	calcQueueNum QueuesNumCalculator,
	calcSize SizeCalculator[T],
) Cache[K, T] {
	c := new(MQ[K, T])
	c.cache = newNtsMqCache[K, T](queues, maxSize, qOutSize, lifeTime, calcQueueNum, calcSize)
	return c
}

func (c *MQ[K, T]) Put(key K, item T) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.put(key, item)
}

func (c *MQ[K, T]) Get(key K, def T) (T, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.cache.get(key, def)
}

func (c *MQ[K, T]) Delete(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.delete(key)
}

type ntsMqCache[K comparable, T any] struct {
	q     []*list.Queue[cacheEntryMQ[K, T]]
	items map[K]*list.Node[cacheEntryMQ[K, T]]

	qOut     *list.Queue[cacheEntryOutMQ[K]]
	itemsOut map[K]*list.Node[cacheEntryOutMQ[K]]

	queues   byte
	maxSize  uint64
	qOutSize uint64
	lifeTime uint64

	calcSize     SizeCalculator[T]
	calcQueueNum QueuesNumCalculator

	currentTime uint64
	currentSize uint64
}

func newNtsMqCache[K comparable, T any](
	queues byte,
	maxSize, qOutSize, lifeTime uint64,
	calcQueueNum QueuesNumCalculator,
	calcSize SizeCalculator[T],
) *ntsMqCache[K, T] {
	if nil == calcSize {
		calcSize = func(T) uint64 { return 1 }
	}
	if nil == calcQueueNum {
		calcQueueNum = func(hits uint64) byte {
			if hits < 1 {
				return 0
			}
			r := math.Log2(float64(hits))
			return byte(r)
		}
	}
	qs := make([]*list.Queue[cacheEntryMQ[K, T]], queues)
	for i := byte(0); i < queues; i++ {
		qs[i] = list.NewQueue[cacheEntryMQ[K, T]]()
	}
	return &ntsMqCache[K, T]{
		items: make(map[K]*list.Node[cacheEntryMQ[K, T]], maxSize),
		q:     qs,

		qOut:     list.NewQueue[cacheEntryOutMQ[K]](),
		itemsOut: make(map[K]*list.Node[cacheEntryOutMQ[K]]),

		queues:   queues,
		maxSize:  maxSize,
		qOutSize: qOutSize,
		lifeTime: lifeTime,

		calcSize:     calcSize,
		calcQueueNum: calcQueueNum,
	}
}

func (c *ntsMqCache[K, T]) adjust() {
	c.currentTime += 1
	for k := byte(1); k < c.queues; k++ {
		e := c.q[k].Head()
		if nil == e {
			continue
		}
		if e.Value().expire < c.currentTime {
			e = c.q[k].Dequeue()
			entry := e.Value()
			entry.expire = c.expire()
			entry.qNum = k - 1
			c.q[k-1].Enqueue(entry)
		}
	}
}

func (c *ntsMqCache[K, T]) evict() {
	for k := byte(0); k < c.queues; k++ {
		victim := c.q[k].Dequeue()
		if nil == victim {
			continue
		}
		key := victim.Value().key
		delete(c.items, key)
		if uint64(c.qOut.Len()) > c.qOutSize {
			drop := c.qOut.Dequeue()
			if drop != nil {
				delete(c.itemsOut, drop.Value().key)
			}
		}
		entry := cacheEntryOutMQ[K]{key: key, hits: victim.Value().hits}
		c.qOut.Enqueue(entry)
		c.itemsOut[key] = c.qOut.Tail()
		size := c.calcSize(victim.Value().value)
		if size > c.currentSize {
			panic("MqCache current size less than size of evicted element")
		}
		c.currentSize -= size
		break
	}
}

func (c *ntsMqCache[K, T]) delete(key K) {
	if e, ok := c.items[key]; ok {
		delete(c.items, key)
		c.q[e.Value().qNum].Remove(e)
		size := c.calcSize(e.Value().value)
		if size > c.currentSize {
			panic("MqCache current size less than size of deleted element")
		}
		c.currentSize -= size
	}
	if e, ok := c.itemsOut[key]; ok {
		delete(c.itemsOut, key)
		c.qOut.Remove(e)
	}
}

func (c *ntsMqCache[K, T]) put(key K, value T) {
	defer c.adjust()
	var curItem cacheEntryMQ[K, T]
	curSize := c.calcSize(value)
	if curSize > c.maxSize {
		panic("MqCache put element larger than cache size")
	}
	e, ok := c.items[key]
	checkMove := false
	curQ := byte(0)
	if ok {
		curItem = e.Value()
		curQ = curItem.qNum
		checkMove = true
	} else {
		curItem.key = key
		if k, ok := c.itemsOut[key]; ok {
			delete(c.itemsOut, key)
			c.qOut.Remove(k)
			curItem.hits = k.Value().hits
		}
	}

	for (curSize + c.currentSize) > c.maxSize {
		c.evict()
	}

	curItem.hits += 1
	curItem.value = value
	curItem.qNum = c.queueNum(curItem.hits)
	curItem.expire = c.expire()

	c.currentSize += curSize

	if checkMove {
		if curQ != curItem.qNum {
			c.q[curQ].Remove(e)
		} else {
			e.SetValue(curItem)
			c.q[curItem.qNum].MoveToBack(e)
			return
		}
	}

	c.q[curItem.qNum].Enqueue(curItem)
	c.items[key] = c.q[curItem.qNum].Tail()
}

func (c *ntsMqCache[K, T]) get(key K, def T) (T, bool) {
	defer c.adjust()
	e, ok := c.items[key]
	if ok {
		curItem := e.Value()
		curQ := curItem.qNum
		curItem.hits += 1
		curItem.qNum = c.queueNum(curItem.hits)
		curItem.expire = c.expire()
		if curQ != curItem.qNum {
			c.q[curQ].Remove(e)
			c.q[curItem.qNum].Enqueue(curItem)
			c.items[key] = c.q[curItem.qNum].Tail()
		} else {
			e.SetValue(curItem)
			c.q[curItem.qNum].MoveToBack(e)
		}
		return curItem.value, true
	}
	return def, false
}

func (c *ntsMqCache[K, T]) queueNum(hits uint64) byte {
	qn := c.calcQueueNum(hits)
	if qn < 0 {
		qn = 0
	} else if qn > c.queues {
		qn = c.queues
	}
	return qn
}

func (c *ntsMqCache[K, T]) expire() uint64 {
	return c.currentTime + c.lifeTime
}

package allcache

import "github.com/satmaelstorm/list"

type fifo[K comparable, T any] struct {
	items map[K]*list.Node[cacheEntry[K, T]]
	queue *list.Queue[cacheEntry[K, T]]
}

func newFifo[K comparable, T any]() *fifo[K, T] {
	return &fifo[K, T]{
		items: make(map[K]*list.Node[cacheEntry[K, T]]),
		queue: list.NewQueue[cacheEntry[K, T]](),
	}
}

func (q *fifo[K, T]) enqueue(key K, value T) {
	q.queue.Enqueue(cacheEntry[K, T]{key, value})
	q.items[key] = q.queue.Tail()
}

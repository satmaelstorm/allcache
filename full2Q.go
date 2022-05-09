package allcache

import "github.com/satmaelstorm/list"

//non thead safe full version 2Q - @see http://www.vldb.org/conf/1994/P439.PDF
type ntsFull2Q[K comparable, T any] struct {
	items map[K]*list.Node[cacheEntry2Q[K, T]]
	am    *list.Queue[cacheEntry2Q[K, T]]
	a1in  *list.Queue[cacheEntry2Q[K, T]]

	itemsOut map[K]struct{}
	a1out    *list.Queue[K]

	amSize    int64
	a1InSize  int64
	a1outSize int64
	totalSize int64
}

func newNtsFull2Q[K comparable, T any](a1InSize, a1OutSize, amSize int64) *ntsFull2Q[K, T] {
	return &ntsFull2Q[K, T]{
		items: make(map[K]*list.Node[cacheEntry2Q[K, T]], amSize+a1InSize),
		am:    list.NewQueue[cacheEntry2Q[K, T]](),
		a1in:  list.NewQueue[cacheEntry2Q[K, T]](),

		itemsOut: make(map[K]struct{}, a1OutSize),
		a1out:    list.NewQueue[K](),

		amSize:    amSize,
		a1InSize:  a1InSize,
		a1OutSize: a1OutSize,

		totalSize: amSize + a1InSize,
	}
}

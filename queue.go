package allcache

import "github.com/satmaelstorm/list"

type queue[K comparable, T any] struct {
	items      map[K]*list.Node[T]
	evictQueue *list.Queue[T]
	lifetime   int64
}

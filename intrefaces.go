package allcache

type Cache[K comparable, T any] interface {
	Put(key K, item T) error
	Get(key K) (T, error)
}

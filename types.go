package allcache

type SizeCalculator[T any] func(T) int64

type Cache[K comparable, T any] interface {
	Put(key K, item T)
	Get(key K, def T) (T, bool)
	Delete(key K)
}

package allcache

type SizeCalculator[T any] func(T) uint64

type QueuesNumCalculator func(hits uint64) byte

type Cache[K comparable, T any] interface {
	Put(key K, item T)
	Get(key K, def T) (T, bool)
	Delete(key K)
}

package allcache

type cacheEntry[K comparable, T any] struct {
	key   K
	value T
}

type cacheEntry2Q[K comparable, T any] struct {
	cacheEntry[K, T]
	isAm bool
}

type cacheEntryMQ[K comparable, T any] struct {
	cacheEntry[K, T]
	qNum   byte
	hits   uint64
	expire uint64
}

type cacheEntryOutMQ[K comparable] struct {
	key  K
	hits uint64
}

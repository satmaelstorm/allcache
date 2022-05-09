package allcache

type cacheEntry[K comparable, T any] struct {
	key   K
	value T
}

type cacheEntry2Q[K comparable, T any] struct {
	cacheEntry[K, T]
	isAm bool
}

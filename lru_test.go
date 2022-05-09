package allcache

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type suiteNtsLRU struct {
	suite.Suite
	cache *ntsLRU[string, int]
}

func TestNtsLru(t *testing.T) {
	suite.Run(t, new(suiteNtsLRU))
}

func (s *suiteNtsLRU) SetupTest() {
	keysStream := []string{"1", "2", "3", "4", "5", "6", "7"}
	valuesStream := []int{1, 2, 3, 4, 5, 6, 7}
	s.cache = newNtsLRU[string, int](5, nil)
	for i := 0; i < len(valuesStream); i++ {
		k := keysStream[i]
		v := valuesStream[i]
		s.cache.put(k, v)
	}
}

func (s *suiteNtsLRU) TestEvictCache() {
	r, ok := s.cache.get("1", -1)
	s.False(ok)
	s.Equal(-1, r)

	r, ok = s.cache.get("2", -2)
	s.False(ok)
	s.Equal(-2, r)

	r, ok = s.cache.get("3", 0)
	s.True(ok)
	s.Equal(3, r)

	r, ok = s.cache.get("4", 0)
	s.True(ok)
	s.Equal(4, r)

	r, ok = s.cache.get("5", 0)
	s.True(ok)
	s.Equal(5, r)

	r, ok = s.cache.get("6", 0)
	s.True(ok)
	s.Equal(6, r)

	r, ok = s.cache.get("7", 0)
	s.True(ok)
	s.Equal(7, r)
}

func (s *suiteNtsLRU) TestDeleteCache() {
	r, ok := s.cache.get("7", 0)
	s.True(ok)
	s.Equal(7, r)

	s.cache.delete("7")

	r, ok = s.cache.get("7", 0)
	s.False(ok)
	s.Equal(0, r)
}

func (s *suiteNtsLRU) TestGetCache() {

	r, ok := s.cache.get("5", 0)
	s.True(ok)
	s.Equal(5, r)

	s.Equal(5, s.cache.evictQueue.Tail().Value().value)

	r, ok = s.cache.get("4", 0)
	s.True(ok)
	s.Equal(4, r)

	s.Equal(4, s.cache.evictQueue.Tail().Value().value)

	r, ok = s.cache.get("key", 0)
	s.False(ok)
	s.Equal(0, r)
}

func (s *suiteNtsLRU) TestPutCache() {
	r, ok := s.cache.get("4", 0)
	s.True(ok)
	s.Equal(4, r)

	s.cache.put("4", 10)

	r, ok = s.cache.get("4", 0)
	s.True(ok)
	s.Equal(10, r)
}

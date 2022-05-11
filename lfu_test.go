package allcache

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type suiteNtsLFU struct {
	suite.Suite
	cache *ntsLFU[string, int]
}

func TestNtsLfu(t *testing.T) {
	suite.Run(t, new(suiteNtsLFU))
}

func (s *suiteNtsLFU) SetupTest() {
	keysStream := []string{"1", "2", "3", "4", "5", "6", "7"}
	valuesStream := []int{1, 2, 3, 4, 5, 6, 7}
	s.cache = newNtsLFU[string, int](5)
	for i := 0; i < len(valuesStream); i++ {
		k := keysStream[i]
		v := valuesStream[i]
		s.cache.put(k, v)
		if i >= 1 && i <= 5 {
			s.cache.get(k, 0)
			if i >= 1 && i <= 4 {
				s.cache.get(k, 0)
			}
		}
	}
}

func (s *suiteNtsLFU) TestEvictCache() {
	r, ok := s.cache.get("1", -1)
	s.False(ok)
	s.Equal(-1, r)

	r, ok = s.cache.get("2", 0)
	s.True(ok)
	s.Equal(2, r)

	r, ok = s.cache.get("3", 0)
	s.True(ok)
	s.Equal(3, r)

	r, ok = s.cache.get("4", 0)
	s.True(ok)
	s.Equal(4, r)

	r, ok = s.cache.get("5", 0)
	s.True(ok)
	s.Equal(5, r)

	r, ok = s.cache.get("6", -1)
	s.False(ok)
	s.Equal(-1, r)

	r, ok = s.cache.get("7", 0)
	s.True(ok)
	s.Equal(7, r)
}

func (s *suiteNtsLFU) TestDeleteCache() {
	r, ok := s.cache.get("7", 0)
	s.True(ok)
	s.Equal(7, r)

	s.cache.delete("7")

	r, ok = s.cache.get("7", 0)
	s.False(ok)
	s.Equal(0, r)
}

func (s *suiteNtsLFU) TestPutCache() {
	r, ok := s.cache.get("4", 0)
	s.True(ok)
	s.Equal(4, r)

	s.cache.put("4", 10)

	r, ok = s.cache.get("4", 0)
	s.True(ok)
	s.Equal(10, r)
}

func (s *suiteNtsLFU) TestTSVersion() {
	c := NewLFU[int, int](3)
	c.Put(1, 1)

	r, ok := c.Get(1, 0)
	s.True(ok)
	s.Equal(1, r)

	c.Delete(1)

	r, ok = c.Get(1, 0)
	s.False(ok)
	s.Equal(0, r)
}

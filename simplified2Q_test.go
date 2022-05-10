package allcache

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type suiteNtsSimplified2Q struct {
	suite.Suite
	cache *ntsSimplified2Q[string, int]
}

func TestNtsSimplified2Q(t *testing.T) {
	suite.Run(t, new(suiteNtsSimplified2Q))
}

func (s *suiteNtsSimplified2Q) SetupTest() {
	s.cache = newNtsSimplified2Q[string, int](3, 2)
	s.cache.put("1", 1)
	s.cache.get("1", 0)
	s.cache.put("2", 2)
	s.cache.get("2", 0)
	s.cache.put("3", 3)
	s.cache.get("3", 0)
	s.cache.put("4", 4)
	s.cache.put("5", 5)
}

func (s *suiteNtsSimplified2Q) TestFill() {
	r, ok := s.cache.get("1", 0)
	s.True(ok)
	s.Equal(1, r)

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

	r, ok = s.cache.get("6", 0)
	s.False(ok)
	s.Equal(0, r)
}

func (s *suiteNtsSimplified2Q) TestEvict() {
	s.cache.put("6", 6)

	r, ok := s.cache.get("4", 0)
	s.False(ok)
	s.Equal(0, r)

	r, ok = s.cache.get("6", 0)
	s.True(ok)
	s.Equal(6, r)

	s.cache.put("7", 7)

	r, ok = s.cache.get("1", 0)
	s.False(ok)
	s.Equal(0, r)
}

func (s *suiteNtsSimplified2Q) TestPut() {
	r, ok := s.cache.get("5", 0)
	s.True(ok)
	s.Equal(5, r)

	s.cache.put("5", 10)

	r, ok = s.cache.get("5", 0)
	s.True(ok)
	s.Equal(10, r)
}

func (s *suiteNtsSimplified2Q) TestTSVersion() {
	c := NewSimplified2Q[int, int](3, 2)
	c.Put(1, 1)

	r, ok := c.Get(1, 0)
	s.True(ok)
	s.Equal(1, r)

	c.Delete(1)

	r, ok = c.Get(1, 0)
	s.False(ok)
	s.Equal(0, r)
}

package allcache

import (
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
)

type suiteNtsFull2Q struct {
	suite.Suite
	cache *ntsFull2Q[string, int]
}

func TestNtsFull2Q(t *testing.T) {
	suite.Run(t, new(suiteNtsFull2Q))
}

func (s *suiteNtsFull2Q) SetupTest() {
	s.cache = newNtsFull2Q[string, int](3, 2, 10)
	s.cache.put("1", 1)
	s.cache.put("2", 2)
	s.cache.put("3", 3)
	for i := 5; i <= 15; i++ {
		s.cache.put(strconv.Itoa(i), i)
	}
	s.cache.put("1", 1)
	s.cache.put("2", 2)
	s.cache.put("3", 3)
}

func (s *suiteNtsFull2Q) TestFill() {
	r, ok := s.cache.get("1", 0)
	s.True(ok)
	s.Equal(1, r)

	r, ok = s.cache.get("2", 0)
	s.True(ok)
	s.Equal(2, r)

	r, ok = s.cache.get("3", 0)
	s.True(ok)
	s.Equal(3, r)

	r, ok = s.cache.get("14", 0)
	s.True(ok)
	s.Equal(14, r)

	r, ok = s.cache.get("15", 0)
	s.True(ok)
	s.Equal(15, r)

	for i := 5; i < 14; i++ {
		r, ok := s.cache.get(strconv.Itoa(i), 0)
		s.False(ok)
		s.Equal(0, r)
	}
}

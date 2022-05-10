package allcache

import (
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
)

type suiteNtsMqCache struct {
	suite.Suite
	cache *ntsMqCache[string, int]
}

func TestMqCache(t *testing.T) {
	suite.Run(t, new(suiteNtsMqCache))
}

func (s *suiteNtsMqCache) SetupTest() {
	s.cache = newNtsMqCache[string, int](8, 5, 5, 5, func(hits uint64) byte {
		if hits < 2 {
			return 0
		}
		if hits > 8 {
			return 8
		}
		return byte(hits - 1)
	}, nil)

	for i := 1; i <= 10; i++ {
		s.cache.put(strconv.Itoa(i), i)
	}
}

func (s *suiteNtsMqCache) TestFill() {
	for i := 1; i <= 5; i++ {
		r, ok := s.cache.get(strconv.Itoa(i), 0)
		s.False(ok)
		s.Equal(0, r)
	}
	for i := 6; i <= 10; i++ {
		r, ok := s.cache.get(strconv.Itoa(i), 0)
		s.True(ok)
		s.Equal(i, r)
	}
}

func (s *suiteNtsMqCache) TestTSVersion() {
	c := NewMQCache[int, int](8, 5, 5, 5, nil, nil)
	c.Put(1, 1)

	r, ok := c.Get(1, 0)
	s.True(ok)
	s.Equal(1, r)

	c.Delete(1)

	r, ok = c.Get(1, 0)
	s.False(ok)
	s.Equal(0, r)
}

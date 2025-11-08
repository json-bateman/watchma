package movie

import (
	"sync"
	"time"
)

type CachingProvider struct {
	inner         Provider
	cache         []Movie
	lastFetched   time.Time
	cacheDuration time.Duration
	mu            sync.Mutex
}

func NewCachingProvider(inner Provider, cacheDuration time.Duration) *CachingProvider {
	return &CachingProvider{
		inner:         inner,
		cacheDuration: cacheDuration,
	}
}

func (c *CachingProvider) FetchMovies() ([]Movie, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if c.cache != nil && now.Sub(c.lastFetched) < c.cacheDuration {
		return c.cache, nil
	}

	movies, err := c.inner.FetchMovies()
	if err != nil {
		return nil, err
	}

	c.cache = movies
	c.lastFetched = now

	return movies, nil
}

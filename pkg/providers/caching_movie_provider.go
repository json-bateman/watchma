package providers

import (
	"sync"
	"time"
	"watchma/pkg/types"
)

type CachingMovieProvider struct {
	inner         MovieProvider
	cache         []types.Movie
	lastFetched   time.Time
	cacheDuration time.Duration
	mu            sync.Mutex
}

func NewCachingMovieProvider(inner MovieProvider, cacheDuration time.Duration) *CachingMovieProvider {
	return &CachingMovieProvider{
		inner:         inner,
		cacheDuration: cacheDuration,
	}
}

func (c *CachingMovieProvider) FetchMovies() ([]types.Movie, error) {
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

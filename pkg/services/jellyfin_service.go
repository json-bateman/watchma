package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"watchma/pkg/types"
)

type JellyfinService struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger

	// Caching fields
	cacheMutex    sync.Mutex
	cachedMovies  *types.JellyfinItems
	lastFetched   time.Time
	cacheDuration time.Duration
}

func NewJellyfinService(apiKey, baseURL string, logger *slog.Logger) *JellyfinService {
	return &JellyfinService{
		apiKey:        apiKey,
		baseURL:       baseURL,
		httpClient:    http.DefaultClient,
		logger:        logger,
		cacheDuration: time.Minute,
	}
}

// FetchJellyfinMovies fetches all movies with caching (1-minute TTL)
func (c *JellyfinService) FetchJellyfinMovies() (*types.JellyfinItems, error) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	now := time.Now()
	if c.cachedMovies != nil && now.Sub(c.lastFetched) < c.cacheDuration {
		return c.cachedMovies, nil
	}

	c.logger.Info("Fetching Jellyfin movies")
	req, err := c.newJellyfinRequest("GET", "/Items?IncludeItemTypes=Movie&Recursive=true")
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result types.JellyfinItems
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	c.cachedMovies = &result
	c.lastFetched = now

	return &result, nil
}

// newJellyfinRequest sets Jellyfin Token as the header on the request.
func (c *JellyfinService) newJellyfinRequest(method, apiEndpoint string) (*http.Request, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("Jellyfin api_key has not been set in settings.json")
	}

	fullURL := fmt.Sprintf("%s%s", c.baseURL, apiEndpoint)
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Emby-Token", c.apiKey)
	return req, nil
}

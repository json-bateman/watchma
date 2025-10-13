package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"watchma/pkg/types"
)

type JellyfinService struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewJellyfinService(apiKey, baseURL string) *JellyfinService {
	return &JellyfinService{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
	}
}

// FetchJellyfinMovies fetches all movies from configured Jellyfin Server
func (c *JellyfinService) FetchJellyfinMovies() (*types.JellyfinItems, error) {
	req, err := c.newJellyfinRequest("GET", "/Items?IncludeItemTypes=Movie&Recursive=true")
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result types.JellyfinItems
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {

		return nil, err
	}

	return &result, nil
}

// newJellyfinRequest sets Jellyfin Token as the header on the request.
func (c *JellyfinService) newJellyfinRequest(method, apiEndpoint string) (*http.Request, error) {
	apiKey := c.apiKey
	baseUrl := c.baseURL
	if apiKey == "" {
		return nil, fmt.Errorf("Jellyfin api_key has not been set in settings.json")
	}

	fullURL := fmt.Sprintf("%s%s", baseUrl, apiEndpoint)
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Emby-Token", apiKey)
	return req, nil
}

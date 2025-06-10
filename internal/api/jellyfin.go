package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/json-bateman/jellyfin-grabber/internal/config"
)

// Sets Jellyfin Token as the header on the request.
func newJellyfinRequest(method, apiEndpoint string) (*http.Request, error) {
	apiKey := config.Config.Jellyfin.APIKey
	baseUrl := config.Config.Jellyfin.BaseURL
	if apiKey == "" {
		slog.Warn("Jellyfin api_key has not been set in settings.json")
		return nil, fmt.Errorf("Jellyfin api_key has not been set in settings.json")
	}

	fullURL := fmt.Sprintf("%s%s", baseUrl, apiEndpoint)
	fmt.Println(apiKey)
	fmt.Println(fullURL)
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		slog.Error("Error Fetching from provided URL")
		return nil, err
	}

	req.Header.Set("X-Emby-Token", apiKey)
	return req, nil
}

func FetchJellyfinMovies() (*JellyfinItems, error) {
	req, err := newJellyfinRequest("get", "/Items?IncludeItemTypes=Movie&Recursive=true")
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	printHtmlRes(resp)

	defer resp.Body.Close()

	var result JellyfinItems
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

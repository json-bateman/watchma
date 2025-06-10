package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

// Sets Jellyfin Token as the header on the request.
func newJellyfinRequest(method, path string) (*http.Request, error) {
	token := os.Getenv("JELLYFIN_TOKEN")
	if token == "" {
		slog.Warn("Jellyfin token has not been set in settings.json")
		return nil, fmt.Errorf("Jellyfin token has not been set in settings.json")
	}

	fullURL := fmt.Sprintf("http://lilnasx.cloud/%s", path)
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		slog.Error("Error Fetching from provided URL")
		return nil, err
	}

	req.Header.Set("X-Emby-Token", os.Getenv("JELLYFIN_TOKEN"))
	return req, nil
}

func FetchJellyfinMovies() (*JellyfinItems, error) {
	req, err := newJellyfinRequest("get", "http://lilnasx.cloud/Items?IncludeItemTypes=Movie&Recursive=true")
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result JellyfinItems
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

package jellyfin

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/json-bateman/jellyfin-grabber/internal/config"
)

// Sets Jellyfin Token as the header on the request.
func newJellyfinRequest(method, apiEndpoint string) (*http.Request, error) {
	apiKey := config.Config.Jellyfin.APIKey
	baseUrl := config.Config.Jellyfin.BaseURL
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

func FetchJellyfinMovies() (*JellyfinItems, error) {
	req, err := newJellyfinRequest("GET", "/Items?IncludeItemTypes=Movie&Recursive=true")
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

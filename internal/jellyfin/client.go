package jellyfin

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
	}
}

func (c *Client) FetchJellyfinMovies() (*JellyfinItems, error) {
	req, err := c.newJellyfinRequest("GET", "/Items?IncludeItemTypes=Movie&Recursive=true")
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

// Sets Jellyfin Token as the header on the request.
func (c *Client) newJellyfinRequest(method, apiEndpoint string) (*http.Request, error) {
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

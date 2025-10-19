package providers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"watchma/pkg/types"
)

type jellyfinItem struct {
	Name            string  `json:"Name"`
	Id              string  `json:"Id"`
	Container       string  `json:"Container"`
	PremiereDate    string  `json:"PremiereDate"`
	CriticRating    int     `json:"CriticRating"`
	CommunityRating float64 `json:"CommunityRating"`
	ProductionYear  int     `json:"ProductionYear"`
	ImageTags       struct {
		Primary string `json:"Primary"`
	} `json:"ImageTags"`
	Genres []string `json:"Genres"`
}

type jellyfinResponse struct {
	Items []jellyfinItem `json:"Items"`
}

type JellyfinMovieProvider struct {
	apiKey     string
	baseUrl    string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewJellyfinMovieProvider(apiKey string, apiUrl string, logger *slog.Logger) *JellyfinMovieProvider {
	return &JellyfinMovieProvider{
		apiKey:     apiKey,
		baseUrl:    apiUrl,
		logger:     logger,
		httpClient: http.DefaultClient,
	}
}

func (p *JellyfinMovieProvider) FetchMovies() ([]types.Movie, error) {
	p.logger.Info("Fetching Jellyfin movies")
	req, err := p.makeRequest("GET", "/Items?IncludeItemTypes=Movie&Recursive=true&Fields=Genres")
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result jellyfinResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	movies := make([]types.Movie, 0, len(result.Items))
	for _, i := range result.Items {
		movies = append(movies, toMovie(i))
	}

	return movies, nil
}

func (p *JellyfinMovieProvider) makeRequest(method string, pathAndQuery string) (*http.Request, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("jellyfin api_key has not been set in settings.json")
	}

	req, err := http.NewRequest(method, p.baseUrl+pathAndQuery, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Emby-Token", p.apiKey)
	return req, nil
}

func toMovie(item jellyfinItem) types.Movie {
	return types.Movie{
		CommunityRating: item.CommunityRating,
		CriticRating:    item.CriticRating,
		Genres:          item.Genres,
		Id:              item.Id,
		Name:            item.Name,
		PremiereDate:    item.PremiereDate,
		PrimaryImageTag: item.ImageTags.Primary,
		ProductionYear:  item.ProductionYear,
	}
}

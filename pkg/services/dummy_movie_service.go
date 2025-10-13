package services

import "watchma/pkg/types"

type DummyMovieService struct{}

func NewDummyMovieService() *DummyMovieService {
	return &DummyMovieService{}
}

func (d *DummyMovieService) FetchJellyfinMovies() (*types.JellyfinItems, error) {
	// Sample movie data - you can add more here
	movies := []types.JellyfinItem{
		{
			Name:            "The Matrix",
			Id:              "movie-1",
			Container:       "mkv",
			PremiereDate:    "1999-03-31T00:00:00Z",
			CriticRating:    88,
			CommunityRating: 8.7,
			RunTimeTicks:    81600000000, // 2h 16m in ticks
			ProductionYear:  1999,
			ImageTags: struct {
				Primary string `json:"Primary"`
				Logo    string `json:"Logo"`
				Thumb   string `json:"Thumb"`
			}{
				Primary: "matrix-poster",
				Logo:    "matrix-logo",
				Thumb:   "matrix-thumb",
			},
			BackdropImageTags: []string{"matrix-backdrop"},
		},
		{
			Name:            "Inception",
			Id:              "movie-2",
			Container:       "mkv",
			PremiereDate:    "2010-07-16T00:00:00Z",
			CriticRating:    87,
			CommunityRating: 8.8,
			RunTimeTicks:    88800000000, // 2h 28m in ticks
			ProductionYear:  2010,
			ImageTags: struct {
				Primary string `json:"Primary"`
				Logo    string `json:"Logo"`
				Thumb   string `json:"Thumb"`
			}{
				Primary: "inception-poster",
				Logo:    "inception-logo",
				Thumb:   "inception-thumb",
			},
			BackdropImageTags: []string{"inception-backdrop"},
		},
		{
			Name:            "Pulp Fiction",
			Id:              "movie-3",
			Container:       "mkv",
			PremiereDate:    "1994-10-14T00:00:00Z",
			CriticRating:    94,
			CommunityRating: 8.9,
			RunTimeTicks:    92400000000, // 2h 34m in ticks
			ProductionYear:  1994,
			ImageTags: struct {
				Primary string `json:"Primary"`
				Logo    string `json:"Logo"`
				Thumb   string `json:"Thumb"`
			}{
				Primary: "pulp-fiction-poster",
				Logo:    "pulp-fiction-logo",
				Thumb:   "pulp-fiction-thumb",
			},
			BackdropImageTags: []string{"pulp-fiction-backdrop"},
		},
		{
			Name:            "The Shawshank Redemption",
			Id:              "movie-4",
			Container:       "mkv",
			PremiereDate:    "1994-09-23T00:00:00Z",
			CriticRating:    91,
			CommunityRating: 9.3,
			RunTimeTicks:    86400000000, // 2h 24m in ticks
			ProductionYear:  1994,
			ImageTags: struct {
				Primary string `json:"Primary"`
				Logo    string `json:"Logo"`
				Thumb   string `json:"Thumb"`
			}{
				Primary: "shawshank-poster",
				Logo:    "shawshank-logo",
				Thumb:   "shawshank-thumb",
			},
			BackdropImageTags: []string{"shawshank-backdrop"},
		},
		{
			Name:            "The Dark Knight",
			Id:              "movie-5",
			Container:       "mkv",
			PremiereDate:    "2008-07-18T00:00:00Z",
			CriticRating:    94,
			CommunityRating: 9.0,
			RunTimeTicks:    91200000000, // 2h 32m in ticks
			ProductionYear:  2008,
			ImageTags: struct {
				Primary string `json:"Primary"`
				Logo    string `json:"Logo"`
				Thumb   string `json:"Thumb"`
			}{
				Primary: "dark-knight-poster",
				Logo:    "dark-knight-logo",
				Thumb:   "dark-knight-thumb",
			},
			BackdropImageTags: []string{"dark-knight-backdrop"},
		},
		{
			Name:            "Forrest Gump",
			Id:              "movie-6",
			Container:       "mkv",
			PremiereDate:    "1994-07-06T00:00:00Z",
			CriticRating:    82,
			CommunityRating: 8.8,
			RunTimeTicks:    86400000000, // 2h 24m in ticks
			ProductionYear:  1994,
			ImageTags: struct {
				Primary string `json:"Primary"`
				Logo    string `json:"Logo"`
				Thumb   string `json:"Thumb"`
			}{
				Primary: "forrest-gump-poster",
				Logo:    "forrest-gump-logo",
				Thumb:   "forrest-gump-thumb",
			},
			BackdropImageTags: []string{"forrest-gump-backdrop"},
		},
		{
			Name:            "Fight Club",
			Id:              "movie-7",
			Container:       "mkv",
			PremiereDate:    "1999-10-15T00:00:00Z",
			CriticRating:    79,
			CommunityRating: 8.8,
			RunTimeTicks:    82800000000, // 2h 18m in ticks
			ProductionYear:  1999,
			ImageTags: struct {
				Primary string `json:"Primary"`
				Logo    string `json:"Logo"`
				Thumb   string `json:"Thumb"`
			}{
				Primary: "fight-club-poster",
				Logo:    "fight-club-logo",
				Thumb:   "fight-club-thumb",
			},
			BackdropImageTags: []string{"fight-club-backdrop"},
		},
		{
			Name:            "Goodfellas",
			Id:              "movie-8",
			Container:       "mkv",
			PremiereDate:    "1990-09-21T00:00:00Z",
			CriticRating:    96,
			CommunityRating: 8.7,
			RunTimeTicks:    88800000000, // 2h 28m in ticks
			ProductionYear:  1990,
			ImageTags: struct {
				Primary string `json:"Primary"`
				Logo    string `json:"Logo"`
				Thumb   string `json:"Thumb"`
			}{
				Primary: "goodfellas-poster",
				Logo:    "goodfellas-logo",
				Thumb:   "goodfellas-thumb",
			},
			BackdropImageTags: []string{"goodfellas-backdrop"},
		},
	}

	return &types.JellyfinItems{Items: movies}, nil
}

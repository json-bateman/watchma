package providers

import "watchma/pkg/types"

type DummyMovieProvider struct{}

func NewDummyMovieProvider() *DummyMovieProvider {
	return &DummyMovieProvider{}
}

func (p *DummyMovieProvider) FetchMovies() ([]types.Movie, error) {
	movies := []types.Movie{
		{
			CommunityRating: 8.7,
			CriticRating:    88,
			Genres:          []string{},
			Id:              "movie-1",
			Name:            "The Matrix",
			OfficialRating:  "R",
			PremiereDate:    "1999-03-31T00:00:00Z",
			PrimaryImageTag: "matrix-poster",
			ProductionYear:  1999,
		},
		{
			CommunityRating: 8.8,
			CriticRating:    87,
			Genres:          []string{},
			Id:              "movie-2",
			Name:            "Inception",
			OfficialRating:  "R",
			PremiereDate:    "2010-07-16T00:00:00Z",
			PrimaryImageTag: "inception-poster",
			ProductionYear:  2010,
		},
		{
			CommunityRating: 8.9,
			CriticRating:    94,
			Genres:          []string{},
			Id:              "movie-3",
			Name:            "Pulp Fiction",
			OfficialRating:  "R",
			PremiereDate:    "1994-10-14T00:00:00Z",
			PrimaryImageTag: "pulp-fiction-poster",
			ProductionYear:  1994,
		},
		{
			CommunityRating: 9.3,
			CriticRating:    91,
			Genres:          []string{},
			Id:              "movie-4",
			Name:            "The Shawshank Redemption",
			OfficialRating:  "R",
			PremiereDate:    "1994-09-23T00:00:00Z",
			PrimaryImageTag: "shawshank-poster",
			ProductionYear:  1994,
		},
		{
			CommunityRating: 9.0,
			CriticRating:    94,
			Genres:          []string{},
			Id:              "movie-5",
			Name:            "The Dark Knight",
			OfficialRating:  "R",
			PremiereDate:    "2008-07-18T00:00:00Z",
			PrimaryImageTag: "dark-knight-poster",
			ProductionYear:  2008,
		},
		{
			CommunityRating: 8.8,
			CriticRating:    82,
			Genres:          []string{},
			Id:              "movie-6",
			Name:            "Forrest Gump",
			OfficialRating:  "R",
			PremiereDate:    "1994-07-06T00:00:00Z",
			PrimaryImageTag: "forrest-gump-poster",
			ProductionYear:  1994,
		},
		{
			CommunityRating: 8.8,
			CriticRating:    79,
			Genres:          []string{},
			Id:              "movie-7",
			Name:            "Fight Club",
			OfficialRating:  "R",
			PremiereDate:    "1999-10-15T00:00:00Z",
			PrimaryImageTag: "fight-club-poster",
			ProductionYear:  1999,
		},
		{
			CommunityRating: 8.7,
			CriticRating:    96,
			Genres:          []string{},
			Id:              "movie-8",
			Name:            "Goodfellas",
			OfficialRating:  "R",
			PremiereDate:    "1990-09-21T00:00:00Z",
			PrimaryImageTag: "goodfellas-poster",
			ProductionYear:  1990,
		},
	}

	return movies, nil
}

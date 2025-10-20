package services

import (
	"fmt"
	"log/slog"
	"math/rand"
	"watchma/pkg/providers"
	"watchma/pkg/types"

	"sort"
	"strings"
)

type MovieSortField string

const (
	SortByName            MovieSortField = "name"
	SortByYear            MovieSortField = "year"
	SortByCriticRating    MovieSortField = "critic"
	SortByCommunityRating MovieSortField = "community"
)

type MovieQuery struct {
	SortBy     MovieSortField
	Descending bool
	Genre      string
	Search     string
}

type MovieService struct {
	movieProvider providers.MovieProvider
	logger        *slog.Logger
}

func NewMovieService(movieProvider providers.MovieProvider, logger *slog.Logger) *MovieService {
	return &MovieService{
		movieProvider: movieProvider,
		logger:        logger,
	}
}

func (ms *MovieService) GetMovies() ([]types.Movie, error) {
	return ms.movieProvider.FetchMovies()
}

func (ms *MovieService) GetShuffledMovies() ([]types.Movie, error) {
	movies, err := ms.GetMovies()
	if err != nil {
		return movies, err
	}

	rand.Shuffle(len(movies), func(i, j int) {
		movies[i], movies[j] = movies[j], movies[i]
	})

	return movies, nil
}

func (ms *MovieService) GetMoviesWithQuery(q MovieQuery) ([]types.Movie, error) {
	movies, err := ms.GetMovies()
	if err != nil {
		return nil, err
	}

	ms.logger.Info("Started query", "sort", q.SortBy, "filter", q.Genre, "search", q.Search)

	movies = filterByGenre(movies, q.Genre)
	ms.logger.Info(fmt.Sprintf("%+v", movies))
	movies = searchByName(movies, q.Search)
	ms.logger.Info(fmt.Sprintf("%+v", movies))
	sortMovies(movies, q.SortBy, q.Descending)
	ms.logger.Info(fmt.Sprintf("%+v", movies))

	return movies, nil
}

func filterByGenre(movies []types.Movie, genre string) []types.Movie {
	if genre == "" {
		return movies
	}
	filtered := make([]types.Movie, 0)
	for _, m := range movies {
		for _, g := range m.Genres {
			if g == genre {
				filtered = append(filtered, m)
				break
			}
		}
	}
	return filtered
}

func searchByName(movies []types.Movie, query string) []types.Movie {
	if query == "" {
		return movies
	}
	query = strings.ToLower(query)
	filtered := make([]types.Movie, 0)
	for _, m := range movies {
		if strings.Contains(strings.ToLower(m.Name), query) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func sortMovies(movies []types.Movie, sortBy MovieSortField, descending bool) {
	sort.Slice(movies, func(i, j int) bool {
		var less bool
		switch sortBy {
		case SortByName:
			less = strings.ToLower(movies[i].Name) < strings.ToLower(movies[j].Name)
		case SortByYear:
			less = movies[i].ProductionYear < movies[j].ProductionYear
		case SortByCriticRating:
			less = movies[i].CriticRating < movies[j].CriticRating
		case SortByCommunityRating:
			less = movies[i].CommunityRating < movies[j].CommunityRating
		default:
			// fallback to name if unknown
			less = strings.ToLower(movies[i].Name) < strings.ToLower(movies[j].Name)
		}
		if descending {
			return !less
		}
		return less
	})
}

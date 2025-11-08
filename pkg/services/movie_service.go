package services

import (
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"time"
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
	// Movie of the day cache
	movieOfTheDay types.Movie
	cachedDay     time.Time
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

	movies = filterByGenre(movies, q.Genre)
	movies = searchByName(movies, q.Search)
	sortMovies(movies, q.SortBy, q.Descending)

	return movies, nil
}

func filterByGenre(movies []types.Movie, genre string) []types.Movie {
	if genre == "" {
		return movies
	}
	filtered := make([]types.Movie, 0)
	for _, m := range movies {
		if slices.Contains(m.Genres, genre) {
			filtered = append(filtered, m)
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
	if sortBy == "" {
		return
	}
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

// GetMovieOfTheDay returns a unique movie for the current UTC day.
//
// If the movie for today has already been selected, it returns the cached movie.
// Otherwise, it fetches the full list of movies from the MovieService,
// picks one deterministically based on the current date (UTC midnight),
// caches it, and returns it.
//
// Returns an error if fetching movies fails or if no movies are available.
func (ms *MovieService) GetMovieOfTheDay() (types.Movie, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if ms.cachedDay.Equal(today) {
		return ms.movieOfTheDay, nil
	}

	movies, err := ms.GetMovies()
	if err != nil {
		return types.Movie{}, fmt.Errorf("failed to fetch movies: %w", err)
	}

	if len(movies) == 0 {
		return types.Movie{}, fmt.Errorf("no movies available")
	}

	// Pick a new movie deterministically with seeded RNG based on today's date
	rng := rand.New(rand.NewSource(today.Unix()))
	index := rng.Intn(len(movies))

	ms.movieOfTheDay = movies[index]
	ms.cachedDay = today

	return ms.movieOfTheDay, nil
}

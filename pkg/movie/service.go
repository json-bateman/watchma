package movie

import (
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"sort"
	"strings"
	"time"
)

type Service struct {
	provider Provider
	logger   *slog.Logger
	// Movie of the day cache
	movieOfTheDay Movie
	cachedDay     time.Time
}

func NewService(provider Provider, logger *slog.Logger) *Service {
	return &Service{
		provider: provider,
		logger:   logger,
	}
}

func (s *Service) GetMovies() ([]Movie, error) {
	return s.provider.FetchMovies()
}

func (s *Service) GetShuffledMovies() ([]Movie, error) {
	movies, err := s.GetMovies()
	if err != nil {
		return movies, err
	}

	rand.Shuffle(len(movies), func(i, j int) {
		movies[i], movies[j] = movies[j], movies[i]
	})

	return movies, nil
}

func (s *Service) GetMoviesWithQuery(q Query) ([]Movie, error) {
	movies, err := s.GetMovies()
	if err != nil {
		return nil, err
	}

	movies = filterByGenre(movies, q.Genre)
	movies = searchByName(movies, q.Search)
	sortMovies(movies, q.SortBy, q.Descending)

	return movies, nil
}

func filterByGenre(movies []Movie, genre string) []Movie {
	if genre == "" {
		return movies
	}
	filtered := make([]Movie, 0)
	for _, m := range movies {
		if slices.Contains(m.Genres, genre) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func searchByName(movies []Movie, query string) []Movie {
	if query == "" {
		return movies
	}
	query = strings.ToLower(query)
	filtered := make([]Movie, 0)
	for _, m := range movies {
		if strings.Contains(strings.ToLower(m.Name), query) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func sortMovies(movies []Movie, sortBy SortField, descending bool) {
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
// Otherwise, it fetches the full list of movies from the Service,
// picks one deterministically based on the current date (UTC midnight),
// caches it, and returns it.
//
// Returns an error if fetching movies fails or if no movies are available.
func (s *Service) GetMovieOfTheDay() (Movie, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if s.cachedDay.Equal(today) {
		return s.movieOfTheDay, nil
	}

	movies, err := s.GetMovies()
	if err != nil {
		return Movie{}, fmt.Errorf("failed to fetch movies: %w", err)
	}

	if len(movies) == 0 {
		return Movie{}, fmt.Errorf("no movies available")
	}

	// Pick a new movie deterministically with seeded RNG based on today's date
	rng := rand.New(rand.NewSource(today.Unix()))
	index := rng.Intn(len(movies))

	s.movieOfTheDay = movies[index]
	s.cachedDay = today

	return s.movieOfTheDay, nil
}

package services

import (
	"fmt"
	"math/rand"
	"time"
	"watchma/pkg/types"
)

type MovieOfTheDayService struct {
	movieService *MovieService
	movie        types.Movie
	day          time.Time
}

func NewMovieOfTheDayService(movieService *MovieService) *MovieOfTheDayService {
	return &MovieOfTheDayService{movieService, types.Movie{}, time.Time{}}
}

// GetMovieOfTheDay returns a unique movie for the current UTC day.
//
// If the movie for today has already been selected, it returns the cached movie.
// Otherwise, it fetches the full list of movies from the MovieService,
// picks one deterministically based on the current date (UTC midnight),
// caches it, and returns it.
//
// Returns an error if fetching movies fails or if no movies are available.
func (motds *MovieOfTheDayService) GetMovieOfTheDay() (types.Movie, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if motds.day.Equal(today) {
		return motds.movie, nil
	}

	movies, err := motds.movieService.GetMovies()
	if err != nil {
		return types.Movie{}, fmt.Errorf("failed to fetch movies: %w", err)
	}

	if len(movies) == 0 {
		return types.Movie{}, fmt.Errorf("no movies available")
	}

	// Pick a new movie deterministically with seeded RNG based on today's date
	rng := rand.New(rand.NewSource(today.Unix()))
	index := rng.Intn(len(movies))

	motds.movie = movies[index]
	motds.day = today

	return motds.movie, nil
}

package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/json-bateman/jellyfin-grabber/internal/types"
)

type MovieOfTheDayService struct {
	movieService ExternalMovieService
	movie        types.JellyfinItem
	day          time.Time
}

func NewMovieOfTheDayService(movieService ExternalMovieService) *MovieOfTheDayService {
	return &MovieOfTheDayService{movieService, types.JellyfinItem{}, time.Time{}}
}

// GetMovieOfTheDay returns a unique movie for the current UTC day.
//
// If the movie for today has already been selected, it returns the cached movie.
// Otherwise, it fetches the full list of movies from the MovieService,
// picks one deterministically based on the current date (UTC midnight),
// caches it, and returns it.
//
// Returns an error if fetching movies fails or if no movies are available.
func (motds *MovieOfTheDayService) GetMovieOfTheDay() (types.JellyfinItem, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if motds.day.Equal(today) {
		return motds.movie, nil
	}

	movies, err := motds.movieService.FetchJellyfinMovies()
	if err != nil {
		return types.JellyfinItem{}, fmt.Errorf("failed to fetch movies: %w", err)
	}

	if len(movies.Items) == 0 {
		return types.JellyfinItem{}, fmt.Errorf("no movies available")
	}

	// Pick a new movie deterministically with seeded RNG based on today's date
	rng := rand.New(rand.NewSource(today.Unix()))
	index := rng.Intn(len(movies.Items))

	motds.movie = movies.Items[index]
	motds.day = today

	return motds.movie, nil
}

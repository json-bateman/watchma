package services

import (
	"math/rand"
	"watchma/pkg/providers"
	"watchma/pkg/types"
)

type MovieService struct {
	movieProvider providers.MovieProvider
}

func NewMovieService(movieProvider providers.MovieProvider) *MovieService {
	return &MovieService{
		movieProvider: movieProvider,
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

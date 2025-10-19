package services

import (
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

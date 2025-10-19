package providers

import "watchma/pkg/types"

type MovieProvider interface {
	FetchMovies() ([]types.Movie, error)
}

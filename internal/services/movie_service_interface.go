package services

import "github.com/json-bateman/jellyfin-grabber/internal/types"

// MovieService interface that both JellyfinService and DummyMovieService implement
// Could possibly extend this interface in the future to support more clients than just Jellyfin
type MovieService interface {
	FetchJellyfinMovies() (*types.JellyfinItems, error)
}


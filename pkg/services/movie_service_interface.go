package services

import "watchma/pkg/types"

// ExternalMovieService interface that both JellyfinService and DummyMovieService implement
// Could possibly extend this interface in the future to support more clients than just Jellyfin
type ExternalMovieService interface {
	FetchJellyfinMovies() (*types.JellyfinItems, error)
}

package services

import "github.com/json-bateman/jellyfin-grabber/internal/types"

// ExternalMovieService interface that both JellyfinService and DummyMovieService implement
// Could possibly extend this interface in the future to support more clients than just Jellyfin
type ExternalMovieService interface {
	FetchJellyfinMovies() (*types.JellyfinItems, error)
}

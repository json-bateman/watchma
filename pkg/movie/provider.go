package movie

// Provider is the interface that movie providers must implement
type Provider interface {
	FetchMovies() ([]Movie, error)
}

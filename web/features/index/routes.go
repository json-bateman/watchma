package index

import (
	"watchma/pkg/movie"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(
	r chi.Router,
	movieService *movie.Service,
) error {
	handlers := newHandlers(movieService)

	r.Get("/", handlers.index)
	r.Get("/shuffle/{number}", handlers.shuffle)
	r.Get("/stats", handlers.stats)

	return nil
}

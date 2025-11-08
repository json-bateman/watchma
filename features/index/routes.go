package index

import (
	"watchma/pkg/services"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(
	r chi.Router,
	movieService *services.MovieService,
) error {
	handlers := newHandlers(movieService)

	r.Get("/", handlers.index)
	r.Get("/shuffle/{number}", handlers.shuffle)

	return nil
}

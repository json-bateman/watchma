package index

import (
	"watchma/db/sqlcgen"
	"watchma/pkg/movie"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(
	r chi.Router,
	movieService *movie.Service,
	queries *sqlcgen.Queries,
) error {
	handlers := newHandlers(movieService, queries)

	r.Get("/", handlers.index)
	r.Get("/shuffle", handlers.shuffle)
	r.Get("/statistics", handlers.stats)
	r.Get("/top5", handlers.top5)

	return nil
}

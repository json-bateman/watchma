package debug

import (
	"watchma/pkg/room"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(
	r chi.Router,
	roomService *room.Service,
) error {
	handlers := newHandlers(roomService)

	r.Get("/debug", handlers.debug)

	return nil
}

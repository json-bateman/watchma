package debug

import (
	"watchma/pkg/services"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(
	r chi.Router,
	roomService *services.RoomService,
) error {
	handlers := newHandlers(roomService)

	r.Get("/debug", handlers.debug)

	return nil
}

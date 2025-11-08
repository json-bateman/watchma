package rooms

import (
	"log/slog"

	"watchma/pkg/services"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

func SetupRoutes(
	r chi.Router,
	roomService *services.RoomService,
	logger *slog.Logger,
	nats *nats.Conn,
) error {
	handlers := newHandlers(roomService, logger, nats)

	r.Get("/host", handlers.host)
	r.Post("/host", handlers.hostForm)
	r.Get("/join", handlers.join)
	r.Get("/sse/join", handlers.joinSSE)

	return nil
}

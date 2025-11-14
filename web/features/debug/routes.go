package debug

import (
	"log/slog"
	"watchma/pkg/room"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

func SetupRoutes(
	r chi.Router,
	roomService *room.Service,
	logger *slog.Logger,
	nc *nats.Conn,
) error {
	handlers := newHandlers(roomService, logger, nc)

	r.Get("/debug", handlers.debug)
	r.Get("/debug/sse", handlers.sse)

	return nil
}

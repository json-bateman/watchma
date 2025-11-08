package auth

import (
	"log/slog"
	"watchma/pkg/services"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(
	r chi.Router,
	authService *services.AuthService,
	logger *slog.Logger,
) error {
	handlers := newHandlers(authService, logger)

	// Public web routes
	r.Get("/login", handlers.Login)
	r.Post("/login", handlers.HandleLogin)
	r.Post("/validate", handlers.ValidatePassword)

	return nil
}

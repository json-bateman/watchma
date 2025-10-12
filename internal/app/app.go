package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/json-bateman/jellyfin-grabber/internal/config"
	"github.com/json-bateman/jellyfin-grabber/internal/handlers/web"
	"github.com/json-bateman/jellyfin-grabber/internal/services"
	"github.com/nats-io/nats.go"
)

type App struct {
	Settings     *config.Settings
	Logger       *slog.Logger
	Router       *chi.Mux
	MovieService services.ExternalMovieService
	NATS         *nats.Conn
}

func New() *App {
	return &App{
		Settings: config.LoadSettings(),
	}
}

func (a *App) Initialize() error {
	a.Logger = config.NewColorLog(a.Settings.LogLevel)
	slog.SetDefault(a.Logger)

	// Use dummy data if Jellyfin credentials aren't provided
	if a.Settings.UseDummyData {
		a.MovieService = services.NewDummyMovieService()
	} else {
		a.MovieService = services.NewJellyfinService(a.Settings.JellyfinApiKey, a.Settings.JellyfinBaseURL)
	}

	a.Router = chi.NewRouter()
	a.Router.Use(middleware.Logger)

	config.SetupFileServer(a.Logger, a.Router)

	roomService := services.NewRoomService()
	movieOfTheDayService := services.NewMovieOfTheDayService(a.MovieService)

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return fmt.Errorf("connect NATS: %w", err)
	}
	// Verify round-trip (ensures we’re really up)
	if err := nc.FlushTimeout(2 * time.Second); err != nil {
		return fmt.Errorf("nats flush failed: %w", err)
	}
	a.NATS = nc
	a.Logger.Info("NATS connected",
		"url", nc.ConnectedUrl(), "server_id", nc.ConnectedServerId(),
		"max_payload", nc.MaxPayload())

	webHandler := web.NewWebHandler(a.Settings, a.MovieService, a.Logger, roomService, movieOfTheDayService, a.NATS)
	webHandler.SetupRoutes(a.Router)

	return nil
}

func (a *App) Run() error {
	a.Logger.Info("Starting server", "port", a.Settings.Port)
	//TODO: maybe a.NATS.Drain() here, need to look into how to gracefully shutdown NATS
	port := fmt.Sprintf(":%d", a.Settings.Port)
	return http.ListenAndServe(port, a.Router)
}

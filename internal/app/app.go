package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/json-bateman/jellyfin-grabber/internal/config"
	"github.com/json-bateman/jellyfin-grabber/internal/handlers/web"
	"github.com/json-bateman/jellyfin-grabber/internal/services"
)

type App struct {
	Settings *config.Settings
	Logger   *slog.Logger
	Router   *chi.Mux
	Jellyfin *services.JellyfinService
	wg       sync.WaitGroup
}

func New() *App {
	return &App{
		Settings: config.LoadSettings(),
	}
}

func (a *App) Initialize() error {
	a.Logger = config.NewColorLog(a.Settings.LogLevel)
	slog.SetDefault(a.Logger)

	a.Jellyfin = services.NewClient(a.Settings.JellyfinApiKey, a.Settings.JellyfinBaseURL)

	a.Router = chi.NewRouter()
	a.Router.Use(middleware.Logger)

	config.SetupFileServer(a.Logger, a.Router)

	roomService := services.NewRoomService()

	webHandler := web.NewWebHandler(a.Settings, a.Jellyfin, a.Logger, roomService)
	webHandler.SetupRoutes(a.Router)

	return nil
}

func (a *App) Run() error {
	a.Logger.Info("Starting server", "port", a.Settings.Port)
	port := fmt.Sprintf(":%d", a.Settings.Port)
	return http.ListenAndServe(port, a.Router)
}

package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/json-bateman/jellyfin-grabber/internal/config"
	"github.com/json-bateman/jellyfin-grabber/internal/jellyfin"
	"github.com/quic-go/webtransport-go"
)

type App struct {
	Config   *config.Config
	Logger   *slog.Logger
	Router   *chi.Mux
	Jellyfin *jellyfin.Client

	// RoomService *services.RoomService
	// UserService *services.UserService

	HTTPServer *http.Server
	WTServer   *webtransport.Server

	shutdown chan os.Signal
	wg       sync.WaitGroup
}

func New() *App {
	return &App{
		Config: config.LoadConfig(),
	}
}

func (a *App) Initialize() error {
	a.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(a.Logger)

	a.Router = chi.NewRouter()
	a.Router.Use(middleware.Logger)

	a.Jellyfin = jellyfin.NewClient(a.Config.JellyfinApiKey, a.Config.JellyfinBaseURL)

	a.setupFileServer()
	a.setupRoutes()

	return nil
}

// File server to serve public assets
func (a *App) setupFileServer() {
	workdir, _ := os.Getwd()
	publicPath := filepath.Join(workdir, "public")
	a.Logger.Info("Setting up file server",
		"working_dir", workdir,
		"public_path", publicPath,
	)

	filesDir := http.Dir(publicPath)
	config.FileServer(a.Router, "/public", filesDir)
	a.Logger.Info("File server configured for /public/*")
}

func (a *App) Run() error {
	a.Logger.Info("Starting server", "port", a.Config.Port)
	port := fmt.Sprintf(":%d", a.Config.Port)
	return http.ListenAndServe(port, a.Router)
}

func (a *App) setupRoutes() {
	// Api
	a.Router.Post("/api/movies", a.PostMovies)
	a.Router.Post("/api/host", a.HostForm)
	a.Router.Post("/api/username", a.SetUsername)

	// Web
	a.Router.Get("/", a.Index)
	a.Router.Get("/host", a.Host)
	a.Router.Get("/join", a.Join)
	a.Router.Get("/room/{roomName}", a.SingleRoom)
	a.Router.Get("/testSSE", a.TestSSE)
	a.Router.Get("/movies", a.Movies)

	// Websocket
	a.Router.Get("/ws/game", a.GameWebSocket)
}

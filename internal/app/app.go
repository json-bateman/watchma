package app

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/json-bateman/jellyfin-grabber/handlers"
	"github.com/json-bateman/jellyfin-grabber/handlers/api"
	"github.com/json-bateman/jellyfin-grabber/handlers/web"
	"github.com/json-bateman/jellyfin-grabber/handlers/websocket"
	"github.com/json-bateman/jellyfin-grabber/internal/config"
	"github.com/quic-go/webtransport-go"
)

type App struct {
	Config *config.Config
	Logger *slog.Logger
	Router *chi.Mux

	// RoomService *services.RoomService
	// UserService *services.UserService
	// JellyfinClient *jellyfin.Client

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
	// Setup logger
	a.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(a.Logger)
	
	// Setup router
	a.Router = chi.NewRouter()
	a.Router.Use(middleware.Logger)
	
	// Setup file server for public assets
	a.setupFileServer()
	
	// Setup routes
	a.setupRoutes()
	
	return nil
}

func (a *App) setupFileServer() {
	workdir, _ := os.Getwd()
	publicPath := filepath.Join(workdir, "public")
	a.Logger.Info("Setting up file server", 
		"working_dir", workdir,
		"public_path", publicPath,
	)
	
	filesDir := http.Dir(publicPath)
	handlers.FileServer(a.Router, "/public", filesDir)
	a.Logger.Info("File server configured for /public/*")
}

func (a *App) Run() error {
	a.Logger.Info("Starting server", "port", 8888)
	return http.ListenAndServe(":8888", a.Router)
}

func (a *App) setupRoutes() {
	// Api
	a.Router.Post("/api/movies", api.PostMovies)
	a.Router.Post("/api/host", api.HostForm)
	a.Router.Post("/api/username", api.SetUsername)

	// Web
	a.Router.Get("/", web.Index)
	a.Router.Get("/host", web.Host)
	a.Router.Get("/join", web.Join)
	a.Router.Get("/room/{roomName}", web.SingleRoom)
	a.Router.Get("/testSSE", web.TestSSE)
	a.Router.Get("/movies", a.Movies)

	// Websocket
	a.Router.Get("/ws/game", websocket.GameWebSocket)
}

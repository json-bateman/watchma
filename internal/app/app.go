package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/json-bateman/jellyfin-grabber/internal/config"
	"github.com/json-bateman/jellyfin-grabber/internal/handlers/api"
	"github.com/json-bateman/jellyfin-grabber/internal/handlers/web"
	"github.com/json-bateman/jellyfin-grabber/internal/services"
	"github.com/nats-io/nats.go"
)

const (
	JOIN_MSG  = "Joined-Room:"
	LEAVE_MSG = "Left-Room:"
)

type App struct {
	Config   *config.Config
	Logger   *slog.Logger
	Router   *chi.Mux
	Jellyfin *config.JellyfinClient
	Nats     *nats.Conn

	HTTPServer *http.Server

	shutdown chan os.Signal
	wg       sync.WaitGroup

	// players / game clients
	gameClients  map[string]map[chan string]bool
	roomMessages map[string][]string
	mu           sync.RWMutex
}

func New() *App {
	return &App{
		Config:       config.LoadConfig(),
		gameClients:  make(map[string]map[chan string]bool),
		roomMessages: make(map[string][]string),
	}
}

func (a *App) Initialize() error {
	a.Logger = config.NewColorLog(a.Config.LogLevel)
	slog.SetDefault(a.Logger)

	a.Jellyfin = config.NewClient(a.Config.JellyfinApiKey, a.Config.JellyfinBaseURL)
	a.Nats = config.NatsConnect(a.Logger)

	a.Router = chi.NewRouter()
	a.Router.Use(middleware.Logger)

	config.SetupFileServer(a.Logger, a.Router)

	roomManager := services.NewRoomManager()

	api := api.NewAPIHandlers(a.Nats, a.gameClients, a.roomMessages, &a.mu)
	a.Router.Route("/api", api.SetupRoutes)

	web := web.NewWebHandlers(a.Config, a.Jellyfin, a.Logger, a.gameClients, a.roomMessages, &a.mu, roomManager)
	a.Router.Route("/", web.SetupRoutes)

	return nil
}

func (a *App) Run() error {
	a.Logger.Info("Starting server", "port", a.Config.Port)
	port := fmt.Sprintf(":%d", a.Config.Port)
	return http.ListenAndServe(port, a.Router)
}

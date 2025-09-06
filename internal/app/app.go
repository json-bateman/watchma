package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/json-bateman/jellyfin-grabber/internal/config"
	"github.com/json-bateman/jellyfin-grabber/internal/handlers/api"
	"github.com/json-bateman/jellyfin-grabber/internal/handlers/web"
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
	Jellyfin *config.Client
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

func (a *App) setupNats() {
	a.Nats = config.Connect()
	a.Nats.Subscribe("chat.*", func(m *nats.Msg) {
		room := strings.TrimPrefix(m.Subject, "chat.")
		message := string(m.Data)

		a.mu.Lock()
		// Store message in room history
		a.roomMessages[room] = append(a.roomMessages[room], message)
		gameClients := a.gameClients[room]
		a.mu.Unlock()

		for gameClient := range gameClients {
			select {
			case gameClient <- message:
			default: // Non-blocking send to prevent deadlock
			}
		}
	})
	a.Nats.Subscribe(JOIN_MSG+".*", func(m *nats.Msg) {
		room := strings.TrimPrefix(m.Subject, JOIN_MSG+".")

		a.mu.Lock()
		gameClients := a.gameClients[room]
		a.mu.Unlock()

		for gameClient := range gameClients {
			select {
			case gameClient <- JOIN_MSG:
			default: // Non-blocking send to prevent deadlock
			}
		}
	})
	a.Nats.Subscribe(LEAVE_MSG+".*", func(m *nats.Msg) {
		room := strings.TrimPrefix(m.Subject, LEAVE_MSG+".")

		a.mu.Lock()
		gameClients := a.gameClients[room]
		a.mu.Unlock()

		for gameClient := range gameClients {
			select {
			case gameClient <- LEAVE_MSG:
			default: // Non-blocking send to prevent deadlock
			}
		}
	})
}

func (a *App) Initialize() error {
	a.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(a.Logger)

	a.Router = chi.NewRouter()
	a.Router.Use(middleware.Logger)

	a.Jellyfin = config.NewClient(a.Config.JellyfinApiKey, a.Config.JellyfinBaseURL)

	a.setupFileServer()
	a.setupNats()

	api := api.NewAPIHandlers(a.Nats, a.gameClients, a.roomMessages, &a.mu)
	a.Router.Route("/api", api.SetupRoutes)

	web := web.NewWebHandlers(a.Config, a.Jellyfin, a.Logger, a.gameClients, a.roomMessages, &a.mu)
	a.Router.Route("/", web.SetupRoutes)

	return nil
}

func (a *App) Run() error {
	a.Logger.Info("Starting server", "port", a.Config.Port)
	port := fmt.Sprintf(":%d", a.Config.Port)
	return http.ListenAndServe(port, a.Router)
}

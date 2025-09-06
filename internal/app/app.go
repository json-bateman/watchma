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
	"github.com/json-bateman/jellyfin-grabber/internal/jellyfin"
	"github.com/json-bateman/jellyfin-grabber/internal/natty"
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
	Jellyfin *jellyfin.Client
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
	a.Nats = natty.Connect()
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

	a.Jellyfin = jellyfin.NewClient(a.Config.JellyfinApiKey, a.Config.JellyfinBaseURL)

	a.setupFileServer()

	// Protected Routes
	a.Router.Group(func(r chi.Router) {
		r.Use(a.RequireUsername)

		// Web
		r.Get("/host", a.Host)
		r.Get("/join", a.Join)
		r.Get("/room/{roomName}", a.SingleRoom)
		r.Get("/message/{room}", a.SingleRoomSSE)
		r.Get("/testSSE", a.TestSSE)
		r.Get("/movies", a.Movies)
		r.Get("/messing", a.Messing)
		r.Get("/", a.Index)
	})

	// Public Routes
	a.Router.Group(func(r chi.Router) {
		// Web
		r.Get("/username", a.Username)
		r.Get("/shuffle/{number}", a.Shuffle)

		// Api
		r.Post("/api/movies", a.PostMovies)
		r.Post("/api/host", a.HostForm)
		r.Post("/api/username", a.SetUsername)
		r.Post("/api/nats/publish", a.PublishToNATS)
		r.Post("/api/rooms/{roomName}/join", a.JoinRoom)
		r.Post("/api/rooms/{roomName}/leave", a.LeaveRoom)

	})

	a.setupNats()

	return nil
}

func (a *App) Run() error {
	a.Logger.Info("Starting server", "port", a.Config.Port)
	port := fmt.Sprintf(":%d", a.Config.Port)
	return http.ListenAndServe(port, a.Router)
}

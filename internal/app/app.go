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

func (a *App) setupRoutes() {
	// Api
	a.Router.Post("/api/movies", a.PostMovies)
	a.Router.Post("/api/host", a.HostForm)
	a.Router.Post("/api/username", a.SetUsername)
	a.Router.Post("/api/nats/publish", a.PublishToNATS)

	// Web
	a.Router.Get("/", a.Index)
	a.Router.Get("/host", a.Host)
	a.Router.Get("/join", a.Join)
	a.Router.Get("/username", a.Username)
	a.Router.Get("/room/{roomName}", a.SingleRoom)
	a.Router.Get("/testSSE", a.TestSSE)
	a.Router.Get("/movies", a.Movies)
	a.Router.Get("/shuffle/{number}", a.Shuffle)
	a.Router.Get("/messing", a.Messing)
	a.Router.Get("/chat/{room}", a.Chat)
	a.Router.Get("/message/{room}", a.ChatSSE)
}

func (a *App) setupNats() {
	a.Nats = natty.Connect()
	a.Nats.Subscribe("chat.*", func(m *nats.Msg) {
		room := strings.TrimPrefix(m.Subject, "chat.")
		message := string(m.Data)

		a.mu.Lock()
		// Store message in room history
		a.roomMessages[room] = append(a.roomMessages[room], message)
		clients := a.gameClients[room]
		a.mu.Unlock()

		for client := range clients {
			select {
			case client <- message:
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
	a.setupRoutes()
	a.setupNats()

	return nil
}

func (a *App) Run() error {
	a.Logger.Info("Starting server", "port", a.Config.Port)
	port := fmt.Sprintf(":%d", a.Config.Port)
	return http.ListenAndServe(port, a.Router)
}

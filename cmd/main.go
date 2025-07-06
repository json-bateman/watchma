package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/handlers"
	"github.com/json-bateman/jellyfin-grabber/handlers/api"
	"github.com/json-bateman/jellyfin-grabber/handlers/pages"
	"github.com/json-bateman/jellyfin-grabber/internal/config"
	"github.com/json-bateman/jellyfin-grabber/internal/log"
)

const PORT = 8080

type AppConfig struct {
	JellyfinApiKey string
	BaseUrl        string
}

func main() {
	// Chi Router for endpoint hits
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// Go Logger for everything else
	sl := log.New(slog.LevelInfo)
	slog.SetDefault(sl)

	// Serve the public/ folder at all times
	workdir, _ := os.Getwd()
	slog.Info(fmt.Sprintf("\nWorking directory:\n%s\n", workdir))
	filesDir := http.Dir(filepath.Join(workdir, "public"))
	handlers.FileServer(r, "/public", filesDir)

	// Load config from environment or .env
	config.LoadConfig()

	// Routes
	r.Get("/", pages.Index)
	r.Get("/movies", pages.Movies)
	r.Get("/host", pages.Host)

	r.Post("/api/movies", api.PostMovies)
	r.Post("/api/host", api.HostForm)

	// Websocket Connection for the game
	r.Get("/ws/game", api.GameWebSocket)

	slog.Info(fmt.Sprintf("\nListening on port :%d\n", PORT))

	err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), r)
	if err != nil {
		slog.Error("server exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

}

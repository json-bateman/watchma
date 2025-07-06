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

func main() {
	// Set up Logger
	sl := log.New(slog.LevelInfo)
	slog.SetDefault(sl)

	// Load jellyfin settings and others into memory at the entrypoint of the app
	err := config.Load("settings.json")
	if err != nil {
		slog.Warn("Could not load settings.json file")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Serve the public/ folder at all times
	workdir, _ := os.Getwd()
	slog.Info(fmt.Sprintf("\nWorking directory:\n%s\n", workdir))
	filesDir := http.Dir(filepath.Join(workdir, "public"))
	handlers.FileServer(r, "/public", filesDir)

	// Routes
	r.Get("/", pages.Index)
	r.Get("/movies", pages.Movies)
	r.Get("/host", pages.Host)
	r.Post("/api/movies", api.PostMovies)

	// Websocket Connection for the game
	r.Get("/ws/game", api.GameWebSocket)

	slog.Info(fmt.Sprintf("\nListening on port :%d\n", PORT))

	err = http.ListenAndServe(fmt.Sprintf(":%d", PORT), r)
	if err != nil {
		slog.Error("server exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

}

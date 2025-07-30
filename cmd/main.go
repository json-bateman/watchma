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
	"github.com/json-bateman/jellyfin-grabber/handlers/web"
	"github.com/json-bateman/jellyfin-grabber/handlers/websocket"
	"github.com/json-bateman/jellyfin-grabber/internal/config"
	"github.com/json-bateman/jellyfin-grabber/internal/log"
)

const PORT = 8888

func main() {
	r := chi.NewRouter()
	// Chi Router Logger for endpoint hits
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

	// Api
	r.Post("/api/movies", api.PostMovies)
	r.Post("/api/host", api.HostForm)
	r.Post("/api/username", api.SetUsername)

	// Web
	r.Get("/", web.Index)
	r.Get("/movies", web.Movies)
	r.Get("/host", web.Host)
	r.Get("/join", web.Join)
	r.Get("/room/{roomName}", web.SingleRoom)
	r.Get("/testSSE", web.TestSSE)

	// Websocket
	r.Get("/ws/game", websocket.GameWebSocket)

	slog.Info(fmt.Sprintf("\nListening on port :%d\n", PORT))

	err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), r)
	if err != nil {
		slog.Error("server exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

}

package router

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	authPkg "watchma/pkg/auth"
	"watchma/pkg/movie"
	"watchma/pkg/openai"
	"watchma/pkg/room"
	"watchma/web"
	"watchma/web/features/auth"
	"watchma/web/features/debug"
	"watchma/web/features/game"
	"watchma/web/features/index"
	"watchma/web/features/rooms"
	"watchma/web/views/http_error"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

type WebHandlerServices struct {
	MovieService   *movie.Service
	RoomService    *room.Service
	AuthService    *authPkg.AuthService
	OpenAiProvider *openai.Provider
}

// WebHandler holds dependencies needed by web handlers
type WebHandler struct {
	jfinBaseUrl string
	jfinApiKey  string
	services    *WebHandlerServices
	logger      *slog.Logger
	NATS        *nats.Conn
}

// NewWebHandler creates a new web handlers instance
func NewWebHandler(jellyfinBaseUrl string, jellyfinApiKey string, logger *slog.Logger, nc *nats.Conn, services *WebHandlerServices) *WebHandler {
	return &WebHandler{
		jfinBaseUrl: jellyfinBaseUrl,
		jfinApiKey:  jellyfinApiKey,
		logger:      logger,
		NATS:        nc,
		services:    services,
	}
}

// Sets up all Web Routes through Chi Router.
// Web Routes should write web elements to http.ResponseWriter (I.E. SSE, HTML, JSON)
func (h *WebHandler) SetupRoutes(r chi.Router) {
	r.Get("/images/{itemId}", proxyJellyfinImage(h.jfinBaseUrl, h.jfinApiKey, h.logger))

	auth.SetupRoutes(r, h.services.AuthService, h.logger)

	// Protected web routes
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireLogin(h.services.AuthService, h.logger))

		index.SetupRoutes(r, h.services.MovieService)
		debug.SetupRoutes(r, h.services.RoomService, h.logger, h.NATS)
		// Room Setup
		rooms.SetupRoutes(r, h.services.RoomService, h.logger, h.NATS)
		// Main Game Loop (lobby, draft, voting, announce)
		game.SetupRoutes(r, h.services.RoomService, h.services.MovieService, h.services.OpenAiProvider, h.logger, h.NATS)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		web.RenderPageNoLayout(http_error.NotFound(), "404-gang", w, r)
	})
}

// proxyJellyfinImage is to allow aggressive caching of jellyfin movie posters.
// Jellyfin by default is no-cache.
func proxyJellyfinImage(jfinBaseUrl string, jfinApiKey string, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		itemId := chi.URLParam(r, "itemId")
		tag := r.URL.Query().Get("tag")
		width := r.URL.Query().Get("width")
		height := r.URL.Query().Get("height")

		if jfinApiKey == "" {
			http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
			logger.Warn("Cannot fetch Image without jellyfinApiKey, Please set JELLYFIN_API_KEY in your environment")
			return
		}

		jellyfinURL := fmt.Sprintf("%s/Items/%s/Images/Primary?tag=%s&width=%s&height=%s",
			jfinBaseUrl, itemId, tag, width, height)

		req, err := http.NewRequest("GET", jellyfinURL, nil)
		if err != nil {
			http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
			logger.Error("Failed to create jellyfin image request", "error", err, "url", jellyfinURL)
			return
		}

		req.Header.Set("X-Emby-Token", jfinApiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
			logger.Error("Failed to fetch jellyfin image", "error", err, "url", jellyfinURL)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("Jellyfin error: %d", resp.StatusCode), http.StatusInternalServerError)
			logger.Warn("Jellyfin returned non-200 status", "status", resp.StatusCode, "url", jellyfinURL)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))

		if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
			w.Header().Set("Last-Modified", lastModified)
		}

		io.Copy(w, resp.Body)
	}
}

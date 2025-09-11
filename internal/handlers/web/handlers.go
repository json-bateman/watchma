package web

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/config"
	"github.com/json-bateman/jellyfin-grabber/internal/services"
	"github.com/json-bateman/jellyfin-grabber/view"
)

// WebHandler holds dependencies needed by web handlers
type WebHandler struct {
	settings    *config.Settings
	jellyfin    *services.JellyfinService
	logger      *slog.Logger
	roomService *services.RoomService
	sseClients  map[string]map[chan string]bool

	mu sync.RWMutex
}

// NewWebHandler creates a new web handlers instance
func NewWebHandler(cfg *config.Settings, jf *services.JellyfinService, l *slog.Logger, rs *services.RoomService) *WebHandler {
	return &WebHandler{
		settings:    cfg,
		jellyfin:    jf,
		logger:      l,
		roomService: rs,
		sseClients:  make(map[string]map[chan string]bool),
	}
}

// Sets up all Web Routes through Chi Router.
// Web Routes should return web elements (I.E. SSE, HTML)
func (h *WebHandler) SetupRoutes(r chi.Router) {
	// Public web routes
	r.Get("/username", h.Username)
	r.Post("/username", h.SetUsername)
	r.Get("/shuffle/{number}", h.Shuffle)
	r.Post("/movies", h.PostMovies)

	// Protected web routes
	r.Group(func(r chi.Router) {
		r.Use(RequireUsername)

		r.Get("/", h.Index)
		r.Get("/host", h.Host)
		r.Post("/host", h.HostForm)
		r.Get("/join", h.Join)
		r.Get("/room/{roomName}", h.SingleRoom)
		r.Get("/message/{room}", h.SingleRoomSSE)
		r.Get("/rooms/{roomName}/movies", h.Movies)
		r.Get("/movies", h.Movies)

		r.Post("/rooms/{roomName}/message", h.PublishChatMessage)
		r.Post("/rooms/{roomName}/join", h.JoinRoom)
		r.Post("/rooms/{roomName}/leave", h.LeaveRoom)
		r.Post("/rooms/{roomName}/ready", h.Ready)
		r.Post("/rooms/{roomName}/start", h.StartGame)
		r.Post("/rooms/{roomName}/submit", h.PublishChatMessage)
	})
}

func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	component := view.IndexPage("Movie Showdown")
	templ.Handler(component).ServeHTTP(w, r)
}

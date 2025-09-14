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
	settings     *config.Settings
	movieService services.MovieService
	logger       *slog.Logger
	roomService  *services.RoomService
	sseClients   map[string]map[chan string]bool

	mu sync.RWMutex
}

// NewWebHandler creates a new web handlers instance
func NewWebHandler(cfg *config.Settings, ms services.MovieService, l *slog.Logger, rs *services.RoomService) *WebHandler {
	return &WebHandler{
		settings:     cfg,
		movieService: ms,
		logger:       l,
		roomService:  rs,
		sseClients:   make(map[string]map[chan string]bool),
	}
}

// Sets up all Web Routes through Chi Router.
// Web Routes should return web elements (I.E. SSE, HTML)
func (h *WebHandler) SetupRoutes(r chi.Router) {
	// Public web routes
	r.Get("/", h.Index)
	r.Get("/username", h.Username)
	r.Get("/shuffle/{number}", h.Shuffle)

	r.Post("/username", h.SetUsername)

	// Protected web routes
	r.Group(func(r chi.Router) {
		r.Use(RequireUsername)

		r.Get("/host", h.Host)
		r.Get("/join", h.Join)
		r.Get("/sse/join", h.JoinSSE)
		r.Get("/room/{roomName}", h.SingleRoom)
		r.Get("/sse/{roomName}", h.SingleRoomSSE)

		r.Post("/host", h.HostForm)
		r.Post("/message", h.PublishChatMessage)
		r.Post("/room/{roomName}/movies", h.SubmitMovies)
		r.Post("/room/{roomName}/join", h.JoinRoom)
		r.Post("/room/{roomName}/leave", h.LeaveRoom)
		r.Post("/room/{roomName}/ready", h.Ready)
		r.Post("/room/{roomName}/start", h.StartGame)
		r.Post("/room/{roomName}", h.SubmitMovies)
	})
}

func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	component := view.IndexPage("Movie Showdown")
	templ.Handler(component).ServeHTTP(w, r)
}

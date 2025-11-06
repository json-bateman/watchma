package web

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"watchma/pkg/config"
	"watchma/pkg/providers"
	"watchma/pkg/services"
	"watchma/view"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

type WebHandlerServices struct {
	MovieService         *services.MovieService
	RoomService          *services.RoomService
	MovieOfTheDayService *services.MovieOfTheDayService
	AuthService          *services.AuthService
	OpenAiProvider       *providers.OpenApiProvider
}

// WebHandler holds dependencies needed by web handlers
type WebHandler struct {
	settings *config.Settings
	services *WebHandlerServices
	logger   *slog.Logger
	NATS     *nats.Conn
}

// NewWebHandler creates a new web handlers instance
func NewWebHandler(settings *config.Settings, logger *slog.Logger, nc *nats.Conn, services *WebHandlerServices) *WebHandler {
	return &WebHandler{
		settings: settings,
		logger:   logger,
		NATS:     nc,
		services: services,
	}
}

// Sets up all Web Routes through Chi Router.
// Web Routes should write web elements to http.ResponseWriter (I.E. SSE, HTML, JSON)
func (h *WebHandler) SetupRoutes(r chi.Router) {
	// Public web routes
	r.Get("/login", h.Login)
	r.Post("/login", h.HandleLogin)
	r.Post("/validate", h.ValidatePassword)

	// Image proxy (public so images load everywhere)
	r.Get("/images/{itemId}", h.ProxyJellyfinImage)

	// Protected web routes
	r.Group(func(r chi.Router) {
		r.Use(h.RequireLogin)

		// Debug endpoint to observe rooms
		r.Get("/debug", h.Debug)

		r.Get("/", h.Index)
		r.Get("/shuffle/{number}", h.Shuffle)

		// Room Setup
		r.Post("/host", h.HostForm)
		r.Get("/host", h.Host)
		r.Get("/join", h.Join)
		r.Get("/sse/join", h.JoinSSE)

		// Lobby
		r.Get("/room/{roomName}", h.SingleRoom)
		r.Get("/sse/{roomName}", h.SingleRoomSSE)
		r.Post("/message", h.PublishChatMessage)
		r.Post("/room/{roomName}/ready", h.Ready)
		r.Post("/room/{roomName}/start", h.StartGame)

		// Draft
		r.Post("/draft/{roomName}/submit", h.DraftSubmit)
		r.Post("/draft/{roomName}/query", h.QueryMovies)
		r.Patch("/draft/{roomName}/{id}", h.ToggleDraftMovie)
		r.Delete("/draft/{roomName}/{id}", h.DeleteFromSelectedMovies)

		// Voting
		r.Post("/voting/{roomName}/submit", h.VotingSubmit)
		r.Patch("/voting/{roomName}/{id}", h.ToggleVotingMovie)

		// Announce
		r.Get("/announce/{roomName}", h.Announce)
	})
}

func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	movieOfTheDay, err := h.services.MovieOfTheDayService.GetMovieOfTheDay()

	if err != nil {
		// TODO: handle case where no movie of the day was found...
		return
	}

	h.RenderPage(view.IndexPage(movieOfTheDay), "Movie Showdown", w, r)
}

func (h *WebHandler) ProxyJellyfinImage(w http.ResponseWriter, r *http.Request) {
	itemId := chi.URLParam(r, "itemId")
	tag := r.URL.Query().Get("tag")
	width := r.URL.Query().Get("width")
	height := r.URL.Query().Get("height")

	jellyfinURL := fmt.Sprintf("%s/Items/%s/Images/Primary?tag=%s&width=%s&height=%s",
		h.settings.JellyfinBaseURL, itemId, tag, width, height)

	resp, err := http.Get(jellyfinURL)
	if err != nil {
		http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))

	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		w.Header().Set("Last-Modified", lastModified)
	}

	io.Copy(w, resp.Body)
}

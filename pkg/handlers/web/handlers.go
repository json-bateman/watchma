package web

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"watchma/pkg/config"
	"watchma/pkg/services"
	"watchma/view"
)

// WebHandler holds dependencies needed by web handlers
type WebHandler struct {
	settings             *config.Settings
	movieService         services.ExternalMovieService
	logger               *slog.Logger
	roomService          *services.RoomService
	movieOfTheDayService *services.MovieOfTheDayService
	authService          *services.AuthService
	NATS                 *nats.Conn
}

// NewWebHandler creates a new web handlers instance
func NewWebHandler(cfg *config.Settings, ms services.ExternalMovieService, l *slog.Logger, rs *services.RoomService, motds *services.MovieOfTheDayService, authSvc *services.AuthService, nc *nats.Conn) *WebHandler {
	return &WebHandler{
		settings:             cfg,
		movieService:         ms,
		logger:               l,
		roomService:          rs,
		movieOfTheDayService: motds,
		authService:          authSvc,
		NATS:                 nc,
	}
}

// Sets up all Web Routes through Chi Router.
// Web Routes should return web elements (I.E. SSE, HTML)
func (h *WebHandler) SetupRoutes(r chi.Router) {
	// Public web routes
	r.Get("/", h.Index)
	r.Get("/login", h.Login)
	r.Get("/username", h.Username)
	r.Get("/shuffle/{number}", h.Shuffle)

	r.Post("/login", h.HandleLogin)
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
		r.Post("/room/{roomName}/ready", h.Ready)
		r.Post("/room/{roomName}/start", h.StartGame)
		r.Post("/room/{roomName}", h.SubmitMovies)
	})
}

func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	movieOfTheDay, err := h.movieOfTheDayService.GetMovieOfTheDay()

	if err != nil {
		// TODO: handle case where no movie of the day was found...
		return
	}

	component := view.IndexPage("Movie Showdown", movieOfTheDay, h.settings.JellyfinBaseURL)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) NatsPublish(subj string, data []byte) error {
	h.logger.Info("NATS publish", "subject", subj, "bytes", len(data))
	return h.NATS.Publish(subj, data)
}

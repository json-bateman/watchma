package web

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"unicode"

	"watchma/pkg/config"
	"watchma/pkg/services"
	"watchma/view"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"
)

type WebHandlerServices struct {
	MovieService         *services.MovieService
	RoomService          *services.RoomService
	MovieOfTheDayService *services.MovieOfTheDayService
	AuthService          *services.AuthService
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
		r.Post("/draft/{roomName}/submit", h.ToggleDraftSubmit)
		r.Post("/draft/{roomName}/query", h.QueryMovies)
		r.Patch("/draft/{roomName}/{id}", h.ToggleSelectedMovie)
		r.Delete("/draft/{roomName}/{id}", h.DeleteFromSelectedMovies)

		// Voting
		r.Post("/voting/{roomName}/submit", h.VotingSubmit)

		// Random Password Validation
		r.Get("/validate", h.Validate)
		r.Post("/validate", h.ValidatePost2)
	})
}

func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	movieOfTheDay, err := h.services.MovieOfTheDayService.GetMovieOfTheDay()

	if err != nil {
		// TODO: handle case where no movie of the day was found...
		return
	}

	response := NewPageResponse(view.IndexPage(movieOfTheDay, h.settings.JellyfinBaseURL), "Movie Showdown")
	h.RenderPage(response, w, r)
}

type Signals struct {
	Password string `json:"password"`
}

type PwRules struct {
	Valid8     bool `json:"valid8"`
	Valid12    bool `json:"valid12"`
	HasNumber  bool `json:"hasNumber"`
	HasSpecial bool `json:"hasSpecial"`
	HasUpper   bool `json:"hasUpper"`
	HasLower   bool `json:"hasLower"`
}

func (h *WebHandler) Validate(w http.ResponseWriter, r *http.Request) {
	templ.Handler(view.Validate2(false)).ServeHTTP(w, r)
}

func isValidPassword(pw string) bool {
	if len(pw) < 8 {
		return false
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(pw)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(pw)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(pw)

	return hasLower && hasUpper && hasNumber
}

func (h *WebHandler) ValidatePost2(w http.ResponseWriter, r *http.Request) {
	var signals Signals
	if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
		h.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}
	sse := datastar.NewSSE(w, r)
	valid := isValidPassword(signals.Password)
	sse.PatchElementTempl(view.Validate2(valid))
}

func (h *WebHandler) ValidatePost(w http.ResponseWriter, r *http.Request) {
	var signals Signals
	if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
		h.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}

	var rules PwRules

	runes := []rune(signals.Password)
	n := len(runes)

	rules.Valid8 = n >= 8
	rules.Valid12 = n >= 12

	for _, r := range runes {
		switch {
		case unicode.IsLower(r):
			rules.HasLower = true
		case unicode.IsUpper(r):
			rules.HasUpper = true
		case unicode.IsDigit(r):
			rules.HasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			rules.HasSpecial = true
		}
	}

	sse := datastar.NewSSE(w, r)
	sse.MarshalAndPatchSignals(rules)
}

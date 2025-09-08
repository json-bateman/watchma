package web

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/config"
	"github.com/json-bateman/jellyfin-grabber/internal/services"
	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/json-bateman/jellyfin-grabber/internal/utils"
	"github.com/json-bateman/jellyfin-grabber/view"
	"github.com/json-bateman/jellyfin-grabber/view/host"
	"github.com/json-bateman/jellyfin-grabber/view/join"
	"github.com/json-bateman/jellyfin-grabber/view/movies"
	"github.com/json-bateman/jellyfin-grabber/view/username"
	"github.com/starfederation/datastar-go/datastar"
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
		r.Post("/message", h.PublishChatMessage)
		r.Post("/rooms/{roomName}/join", h.JoinRoom)
		r.Post("/rooms/{roomName}/leave", h.LeaveRoom)
		r.Get("/movies", h.Movies)
	})
}

func (h *WebHandler) SetUsername(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")

	http.SetCookie(w, &http.Cookie{
		Name:   "jelly_user",
		Value:  username,
		Path:   "/",
		MaxAge: 30 * 24 * 60 * 60, // 30 days
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	component := view.IndexPage("Sup wit it")
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) Host(w http.ResponseWriter, r *http.Request) {
	component := host.HostPage("username")
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) Join(w http.ResponseWriter, r *http.Request) {
	component := join.JoinPage(h.roomService.Rooms)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) Username(w http.ResponseWriter, r *http.Request) {
	component := username.UsernameForm()
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) Movies(w http.ResponseWriter, r *http.Request) {
	items, err := h.jellyfin.FetchJellyfinMovies()
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "Unable to load movies", http.StatusInternalServerError)
		return
	}

	if items == nil || len(items.Items) == 0 {
		log.Printf("no movies found")
	}

	// TODO: Improve this ranomizer for number of items after room is made
	rand.Shuffle(len(items.Items), func(i, j int) {
		items.Items[i], items.Items[j] = items.Items[j], items.Items[i]
	})

	var randMovies []types.JellyfinItem
	if len(items.Items) >= 8 {
		randMovies = items.Items[:8]
	} else {
		randMovies = items.Items
	}

	component := movies.MoviesPage(randMovies, h.settings.JellyfinBaseURL)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) Shuffle(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	intVal, err := strconv.Atoi(number)
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "param must be a number", http.StatusBadRequest)
		return
	}

	items, err := h.jellyfin.FetchJellyfinMovies()
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "Unable to load movies", http.StatusInternalServerError)
		return
	}

	if items == nil || len(items.Items) == 0 {
		log.Printf("no movies found")
	}

	rand.Shuffle(len(items.Items), func(i, j int) {
		items.Items[i], items.Items[j] = items.Items[j], items.Items[i]
	})

	var randMovies []types.JellyfinItem
	if len(items.Items) >= intVal {
		randMovies = items.Items[:intVal]
	} else {
		randMovies = items.Items
	}

	component := movies.Shuffle(randMovies, h.settings.JellyfinBaseURL)
	templ.Handler(component).ServeHTTP(w, r)
}

func TestSSE(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	for {
		select {
		case <-r.Context().Done():
			return
		default:
			sse.Send("ping", []string{`<div class="text-blue-300">yolo</div>`})
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (h *WebHandler) HostForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	roomName := r.FormValue("roomName")
	moviesStr := r.FormValue("movies")
	maxPlayersStr := r.FormValue("maxplayers")

	movies, err := strconv.Atoi(moviesStr)
	maxPlayers, err := strconv.Atoi(maxPlayersStr)
	if err != nil {
		http.Error(w, "Movies must be a number", http.StatusBadRequest)
		return
	}
	if h.roomService.RoomExists(roomName) {
		http.Error(w, "This room name already exists", http.StatusBadRequest)
		return
	}
	h.roomService.AddRoom(roomName, &types.GameSession{
		MovieNumber: movies,
		MaxPlayers:  maxPlayers,
	})

	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusSeeOther)
}

func (h *WebHandler) PostMovies(w http.ResponseWriter, r *http.Request) {
	var moviesReq types.MovieReq
	fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&moviesReq); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}
	if len(moviesReq.MoviesReq) == 0 {
		utils.WriteJSONError(w, http.StatusBadRequest, "Must include at least 1 movie id.")
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(movies.SubmitButton())

	fmt.Println(moviesReq.MoviesReq)
}

func (h *WebHandler) AddClient(roomName string, client chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.sseClients[roomName] == nil {
		h.sseClients[roomName] = make(map[chan string]bool)
	}
	h.sseClients[roomName][client] = true
}

func (h *WebHandler) RemoveClient(roomName string, client chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.sseClients[roomName], client)
	if len(h.sseClients[roomName]) == 0 {
		delete(h.sseClients, roomName)
	}
}

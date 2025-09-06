package handlers

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
	"github.com/json-bateman/jellyfin-grabber/view/messing"
	"github.com/json-bateman/jellyfin-grabber/view/movies"
	"github.com/json-bateman/jellyfin-grabber/view/rooms"
	"github.com/json-bateman/jellyfin-grabber/view/username"
	"github.com/starfederation/datastar-go/datastar"
)

// WebHandlers holds dependencies needed by web handlers
type WebHandlers struct {
	Config       *config.Config
	Jellyfin     *config.Client
	Logger       *slog.Logger
	gameClients  map[string]map[chan string]bool
	roomMessages map[string][]string
	mu           *sync.RWMutex
}

// NewWebHandlers creates a new web handlers instance
func NewWebHandlers(cfg *config.Config, jf *config.Client, l *slog.Logger, gameClients map[string]map[chan string]bool, roomMessages map[string][]string, mu *sync.RWMutex) *WebHandlers {
	return &WebHandlers{
		Config:       cfg,
		Jellyfin:     jf,
		Logger:       l,
		gameClients:  gameClients,
		roomMessages: roomMessages,
		mu:           mu,
	}
}

// Sets up all Web Routes through Chi Router.
// Web Routes should return web elements (I.E. SSE, HTML)
func (h *WebHandlers) SetupRoutes(r chi.Router) {
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
		r.Get("/movies", h.Movies)
		r.Get("/room/{roomName}", h.SingleRoom)
		r.Get("/message/{room}", h.SingleRoomSSE)
	})
}

func (h *WebHandlers) SetUsername(w http.ResponseWriter, r *http.Request) {
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

func (h *WebHandlers) Index(w http.ResponseWriter, r *http.Request) {
	component := view.IndexPage("Sup wit it")
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandlers) Host(w http.ResponseWriter, r *http.Request) {
	component := host.HostPage("username")
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandlers) Join(w http.ResponseWriter, r *http.Request) {
	component := join.JoinPage(services.AllRooms.Rooms)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandlers) Username(w http.ResponseWriter, r *http.Request) {
	component := username.UsernameForm()
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandlers) SingleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")

	myRoom, ok := services.AllRooms.GetRoom(roomName)
	if !ok {
		component := rooms.NoRoom(roomName)
		templ.Handler(component).ServeHTTP(w, r)
		return
	}

	if myRoom.Game.MaxPlayers <= len(myRoom.Users) {
		component := rooms.RoomFull(roomName)
		templ.Handler(component).ServeHTTP(w, r)
		return
	}

	component := rooms.SingleRoom(myRoom)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandlers) SingleRoomSSE(w http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")
	sse := datastar.NewSSE(w, r)
	client := make(chan string, 100)

	h.mu.Lock()
	// Register for chat updates
	if h.gameClients[room] == nil {
		h.gameClients[room] = make(map[chan string]bool)
	}
	h.gameClients[room][client] = true

	// Send existing message history to new client
	if roomHistory := h.roomMessages[room]; len(roomHistory) > 0 {
		chat := rooms.ChatBox(h.roomMessages[room])
		if err := sse.PatchElementTempl(chat); err != nil {
			h.Logger.Error("Error Patching chatbox on load")
			h.mu.Unlock()
			return
		}
	}

	// Send existing user list to new client
	myRoom, ok := services.AllRooms.GetRoom(room)
	if ok {
		userBox := rooms.UserBox(myRoom)
		if err := sse.PatchElementTempl(userBox); err != nil {
			h.Logger.Error("Error patching initial user list")
		}
	}
	h.mu.Unlock()

	// Cleanup when connection closes
	defer func() {
		h.mu.Lock()
		delete(h.gameClients[room], client)
		if len(h.gameClients[room]) == 0 {
			delete(h.gameClients, room)
		}
		h.mu.Unlock()
		close(client)
	}()

	for {
		select {
		case message := <-client:
			var roomMsg types.RoomMessage
			// Should be able to parse into RoomMessage
			if err := json.Unmarshal([]byte(message), &roomMsg); err == nil {
				switch roomMsg.Subject {
				case utils.JOIN_MSG:
					myRoom, ok := services.AllRooms.GetRoom(room)
					if ok {
						myRoom.AddUser(roomMsg.Username)
						userBox := rooms.UserBox(myRoom)
						if err := sse.PatchElementTempl(userBox); err != nil {
							fmt.Println("Error patching user list")
							return
						}
					}
				case utils.LEAVE_MSG:
					myRoom, ok := services.AllRooms.GetRoom(room)
					if ok {
						myRoom.RemoveUser(roomMsg.Username)
						userBox := rooms.UserBox(myRoom)
						if err := sse.PatchElementTempl(userBox); err != nil {
							fmt.Println("Error patching user list")
							return
						}
					}
				default:
					// Everything else is a chat message
					h.mu.RLock()
					currentMessages := make([]string, len(h.roomMessages[room]))
					copy(currentMessages, h.roomMessages[room])
					h.mu.RUnlock()

					chat := rooms.ChatBox(currentMessages)
					if err := sse.PatchElementTempl(chat); err != nil {
						fmt.Println("Error patching chat message")
						return
					}
				}
			} // else {} What to do if message can't be parsed?
		case <-r.Context().Done():
			return
		}
	}
}

func (h *WebHandlers) Movies(w http.ResponseWriter, r *http.Request) {
	items, err := h.Jellyfin.FetchJellyfinMovies()
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

	component := movies.MoviesPage(randMovies, h.Config.JellyfinBaseURL)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandlers) Shuffle(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	intVal, err := strconv.Atoi(number)
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "param must be a number", http.StatusBadRequest)
		return
	}

	items, err := h.Jellyfin.FetchJellyfinMovies()
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

	component := movies.Shuffle(randMovies, h.Config.JellyfinBaseURL)
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

func Messing(w http.ResponseWriter, r *http.Request) {
	component := messing.Test()
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandlers) HostForm(w http.ResponseWriter, r *http.Request) {
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
	if services.AllRooms.RoomExists(roomName) {
		http.Error(w, "This room name already exists", http.StatusBadRequest)
		return
	}
	services.AllRooms.AddRoom(roomName, &services.GameSession{
		MovieNumber: movies,
		MaxPlayers:  maxPlayers,
	})

	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusSeeOther)
}

func (h *WebHandlers) PostMovies(w http.ResponseWriter, r *http.Request) {
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

package app

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/jellyfin"
	"github.com/json-bateman/jellyfin-grabber/internal/services"
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

type RoomMessage struct {
	Subject  string `json:"subject"`
	Message  string `json:"message"`
	Username string `json:"username"`
}

// RequireUsername middleware checks for jelly_user cookie
func (a *App) RequireUsername(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := utils.GetUsernameFromCookie(r)
		if username == "" {
			http.Redirect(w, r, "/username", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- view ---//
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	component := view.IndexPage("Sup wit it")
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/host ---//
func (a *App) Host(w http.ResponseWriter, r *http.Request) {
	username := utils.GetUsernameFromCookie(r)

	component := host.HostPage(username)
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/shuffle ---//
func (a *App) Shuffle(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	intVal, err := strconv.Atoi(number)
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "param must be a number", http.StatusBadRequest)
		return
	}

	items, err := a.Jellyfin.FetchJellyfinMovies()
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

	var randMovies []jellyfin.JellyfinItem
	if len(items.Items) >= intVal {
		randMovies = items.Items[:intVal]
	} else {
		randMovies = items.Items // fallback if fewer than 8 items
	}

	component := movies.Shuffle(randMovies, a.Config.JellyfinBaseURL)
	templ.Handler(component).ServeHTTP(w, r)
}

func (a *App) Movies(w http.ResponseWriter, r *http.Request) {

	items, err := a.Jellyfin.FetchJellyfinMovies()
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

	var randMovies []jellyfin.JellyfinItem
	if len(items.Items) >= 8 {
		randMovies = items.Items[:8]
	} else {
		randMovies = items.Items // fallback if fewer than 8 items
	}

	component := movies.MoviesPage(randMovies, a.Config.JellyfinBaseURL)
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/join ---//
func (a *App) Join(w http.ResponseWriter, r *http.Request) {
	component := join.JoinPage(services.AllRooms.Rooms)
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/username ---//
func (a *App) Username(w http.ResponseWriter, r *http.Request) {
	component := username.UsernameForm()
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/messing ---//
func (a *App) Messing(w http.ResponseWriter, r *http.Request) {
	component := messing.Test()
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/testSSE ---//
func (a *App) TestSSE(w http.ResponseWriter, r *http.Request) {
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

// --- view/rooms ---//
func (a *App) SingleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")

	myRoom, ok := services.AllRooms.GetRoom(roomName)
	if !ok {
		http.Redirect(w, r, "/join", http.StatusSeeOther)
	}

	component := rooms.SingleRoom(myRoom)
	templ.Handler(component).ServeHTTP(w, r)
}

func (a *App) SingleRoomSSE(w http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")
	sse := datastar.NewSSE(w, r)
	client := make(chan string, 100)

	a.mu.Lock()
	// Register for chat updates
	if a.gameClients[room] == nil {
		a.gameClients[room] = make(map[chan string]bool)
	}
	a.gameClients[room][client] = true

	// Send existing message history to new client
	if roomHistory := a.roomMessages[room]; len(roomHistory) > 0 {
		chat := rooms.ChatBox(a.roomMessages[room])
		if err := sse.PatchElementTempl(chat); err != nil {
			a.Logger.Error("Error Patching chatbox on load")
			a.mu.Unlock()
			return
		}
	}

	// Send existing user list
	myRoom, ok := services.AllRooms.GetRoom(room)
	if ok {
		userBox := rooms.UserBox(myRoom)
		if err := sse.PatchElementTempl(userBox); err != nil {
			a.Logger.Error("Error patching initial user list")
		}
	}
	a.mu.Unlock()

	// Cleanup when connection closes
	defer func() {
		a.mu.Lock()
		delete(a.gameClients[room], client)
		if len(a.gameClients[room]) == 0 {
			delete(a.gameClients, room)
		}
		a.mu.Unlock()
		close(client)
	}()

	for {
		select {
		case message := <-client:
			// Try to parse as JSON first
			var roomMsg RoomMessage
			if err := json.Unmarshal([]byte(message), &roomMsg); err == nil {
				// Handle structured messages
				switch roomMsg.Subject {
				case JOIN_MSG:
					myRoom, ok := services.AllRooms.GetRoom(room)
					if ok {
						myRoom.AddUser(roomMsg.Username)
						userBox := rooms.UserBox(myRoom)
						if err := sse.PatchElementTempl(userBox); err != nil {
							fmt.Println("Error patching user list")
							return
						}
					}
				case LEAVE_MSG:
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
					// Handle old-style chat messages (fallback)
					a.mu.RLock()
					currentMessages := make([]string, len(a.roomMessages[room]))
					copy(currentMessages, a.roomMessages[room])
					a.mu.RUnlock()

					chat := rooms.ChatBox(currentMessages)
					if err := sse.PatchElementTempl(chat); err != nil {
						fmt.Println("Error patching chat message")
						return
					}
				}
			}

		case <-r.Context().Done():
			return
		}
	}
}

func (a *App) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	username := utils.GetUsernameFromCookie(r)

	if username == "" {
		http.Error(w, "No username found", http.StatusBadRequest)
		return
	}

	// Create structured message
	msg := RoomMessage{
		Subject:  LEAVE_MSG,
		Username: username,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	// Broadcast to all clients in this room
	a.mu.RLock()
	if clients, ok := a.gameClients[roomName]; ok {
		for client := range clients {
			select {
			case client <- string(msgBytes):
			default:
				// Client buffer full, skip
			}
		}
	}
	a.mu.RUnlock()

	w.WriteHeader(http.StatusOK)
}

func (a *App) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	username := utils.GetUsernameFromCookie(r)

	if username == "" {
		http.Error(w, "No username found", http.StatusBadRequest)
		return
	}

	// Create structured message
	msg := RoomMessage{
		Subject:  JOIN_MSG,
		Username: username,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	// Broadcast to all clients in this room
	a.mu.RLock()
	if clients, ok := a.gameClients[roomName]; ok {
		for client := range clients {
			select {
			case client <- string(msgBytes):
			default:
				// Client buffer full, skip
			}
		}
	}
	a.mu.RUnlock()

	w.WriteHeader(http.StatusOK)
}

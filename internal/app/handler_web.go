package app

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/game"
	"github.com/json-bateman/jellyfin-grabber/internal/jellyfin"
	"github.com/json-bateman/jellyfin-grabber/view"
	"github.com/json-bateman/jellyfin-grabber/view/chat"
	"github.com/json-bateman/jellyfin-grabber/view/host"
	"github.com/json-bateman/jellyfin-grabber/view/join"
	"github.com/json-bateman/jellyfin-grabber/view/messing"
	"github.com/json-bateman/jellyfin-grabber/view/movies"
	"github.com/json-bateman/jellyfin-grabber/view/rooms"
	"github.com/json-bateman/jellyfin-grabber/view/username"
	"github.com/starfederation/datastar-go/datastar"
)

// --- view ---//
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	component := view.IndexPage("Sup wit it")
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/host ---//
func (a *App) Host(w http.ResponseWriter, r *http.Request) {
	j, _ := r.Cookie("jelly_user")
	component := host.HostPage(j.Value)
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

// --- view/rooms ---//
func (a *App) SingleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")

	var myRoom *game.Room
	for a, b := range game.AllRooms.Rooms {
		if roomName == a {
			myRoom = b
		}
	}

	component := rooms.SingleRoom(myRoom)
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/join ---//
func (a *App) Join(w http.ResponseWriter, r *http.Request) {
	// Check if user has username cookie
	_, err := r.Cookie("jelly_user")
	if err != nil {
		// No username cookie, redirect to username form
		http.Redirect(w, r, "/username", http.StatusSeeOther)
		return
	}

	// User has username, show join page
	component := join.JoinPage(game.AllRooms.Rooms)
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

func (a *App) Chat(w http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")
	component := chat.Chat(room)
	templ.Handler(component).ServeHTTP(w, r)
}

// TODO: maybe change wildcard matching on chat rooms
func (a *App) ChatSSE(w http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")
	sse := datastar.NewSSE(w, r)
	client := make(chan string, 100) // Buffered channel to prevent blocking

	// Add client to room with proper synchronization
	a.mu.Lock()
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
		case <-client:
			// Get current room messages and send to client
			a.mu.RLock()
			currentMessages := make([]string, len(a.roomMessages[room]))
			copy(currentMessages, a.roomMessages[room])
			a.mu.RUnlock()

			chat := rooms.ChatBox(currentMessages)

			if err := sse.PatchElementTempl(chat); err != nil {
				fmt.Println("Error patching message from client")
				return
			}
			// if err := sse.MarshalAndPatchSignals(map[string][]string{
			// 	"message": currentMessages,
			// }); err != nil {
			// 	fmt.Println("err2")
			// 	return
			// }
		case <-r.Context().Done():
			return
		}
	}
}

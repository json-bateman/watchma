package app

import (
	"log"
	"log/slog"
	"math/rand"
	"net/http"
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
	"github.com/starfederation/datastar-go/datastar"
)

// --- view ---//
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	component := view.IndexPage("Sup wit it")
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/host ---//
func (a *App) Host(w http.ResponseWriter, r *http.Request) {
	component := host.HostPage()
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/movies ---//
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
	component := join.JoinPage(game.AllRooms.Rooms)
	templ.Handler(component).ServeHTTP(w, r)
}

// --- view/messing ---//
func (a *App) Messing(w http.ResponseWriter, r *http.Request) {
	component := messing.Test()
	templ.Handler(component).ServeHTTP(w, r)
}

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
	component := chat.Chat()
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
		if err := sse.MarshalAndPatchSignals(map[string][]string{
			"message": roomHistory,
		}); err != nil {
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

			if err := sse.MarshalAndPatchSignals(map[string][]string{
				"message": currentMessages,
			}); err != nil {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

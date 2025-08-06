package app

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/game"
	"github.com/json-bateman/jellyfin-grabber/view"
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

	component := movies.MoviesPage(items, a.Config.JellyfinBaseURL)
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
			time.Sleep(200 * time.Millisecond)
		}
	}
}

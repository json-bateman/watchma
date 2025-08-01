package app

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/json-bateman/jellyfin-grabber/internal/jellyfin"
	"github.com/json-bateman/jellyfin-grabber/view/movies"
)

func (a *App) Movies(w http.ResponseWriter, r *http.Request) {

	allMovies, err := jellyfin.FetchJellyfinMovies(a.Config)
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "Unable to load movies", http.StatusInternalServerError)
		return
	}

	items, err := jellyfin.FetchJellyfinMovies(a.Config)
	if err != nil {
		slog.Error("fetch failed: %v\n" + err.Error())
		http.Error(w, "Unable to load movies", http.StatusInternalServerError)
		return
	}
	if items == nil || len(items.Items) == 0 {
		log.Printf("no movies found")
	}

	component := movies.MoviesPage(allMovies, a.Config.JellyfinBaseURL)
	templ.Handler(component).ServeHTTP(w, r)
}

package handlers

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/json-bateman/jellyfin-grabber/internal/api"
	"github.com/json-bateman/jellyfin-grabber/view/home"
)

type HomeHandler struct{}

func (h HomeHandler) Show(w http.ResponseWriter, r *http.Request) {

	allMovies, err := api.FetchJellyfinMovies()
	if err != nil {
		fmt.Println("Error fetching jellyfin movies!")
	}

	items, err := api.FetchJellyfinMovies()
	if err != nil {
		log.Printf("fetch failed: %v", err)
		http.Error(w, "Unable to load movies", http.StatusInternalServerError)
		return
	}
	if items == nil || len(items.Items) == 0 {
		log.Printf("no movies found")
	}

	for _, m := range allMovies.Items {
		slog.Info(m.Name)
	}

	component := home.Home()
	templ.Handler(component).ServeHTTP(w, r)
}

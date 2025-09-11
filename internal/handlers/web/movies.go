package web

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/json-bateman/jellyfin-grabber/internal/utils"
	"github.com/json-bateman/jellyfin-grabber/view/movies"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) Movies(w http.ResponseWriter, r *http.Request) {
	items, err := h.jellyfin.FetchJellyfinMovies()
	if err != nil {
	}

	if items == nil || len(items.Items) == 0 {
		log.Printf("no movies found")
	}

	// TODO: Improve this ranomizer for number of items after room is made
	rand.Shuffle(len(items.Items), func(i, j int) {
		items.Items[i], items.Items[j] = items.Items[j], items.Items[i]
	})

	var randMovies []types.JellyfinItem
	if len(items.Items) >= 20 {
		randMovies = items.Items[:20]
	} else {
		randMovies = items.Items
	}

	component := movies.MoviesPage(randMovies, h.settings.JellyfinBaseURL, nil)
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

func (h *WebHandler) PostMovies(w http.ResponseWriter, r *http.Request) {
	var moviesReq types.MovieRequest
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

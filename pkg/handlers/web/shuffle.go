package web

import (
	"net/http"
	"strconv"
	"watchma/view/shuffle"

	"github.com/go-chi/chi/v5"
)

// Shuffle returns a page with a shuffled list of movies, up to the number requested in the query parameters
func (h *WebHandler) Shuffle(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	numberOfMovies, err := strconv.Atoi(number)
	if err != nil {
		http.Error(w, "param must be a number", http.StatusBadRequest)
		return
	}

	shuffledMovies, err := h.services.MovieService.GetShuffledMovies()
	if err != nil {
		http.Error(w, "failed to get movies", http.StatusInternalServerError)
		return
	}

	if len(shuffledMovies) > numberOfMovies {
		shuffledMovies = shuffledMovies[:numberOfMovies]
	}

	response := NewPageResponse(shuffle.Shuffle(shuffledMovies, h.settings.JellyfinBaseURL), "Movies")
	h.RenderPage(response, w, r)
}

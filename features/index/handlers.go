package index

import (
	"net/http"
	"strconv"
	"watchma/features/index/pages"
	"watchma/pkg/services"
	"watchma/pkg/web"

	"github.com/go-chi/chi/v5"
)

type handlers struct {
	movieService *services.MovieService
}

func newHandlers(ms *services.MovieService) *handlers {
	return &handlers{
		movieService: ms,
	}
}

func (h *handlers) index(w http.ResponseWriter, r *http.Request) {
	movieOfTheDay, err := h.movieService.GetMovieOfTheDay()

	if err != nil {
		// TODO: handle case where no movie of the day was found...
		return
	}

	web.RenderPage(pages.IndexPage(movieOfTheDay), "Movie Showdown", w, r)
}

// Shuffle returns a page with a shuffled list of movies, up to the number requested in the query parameters
func (h *handlers) shuffle(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	numberOfMovies, err := strconv.Atoi(number)
	if err != nil {
		http.Error(w, "param must be a number", http.StatusBadRequest)
		return
	}

	shuffledMovies, err := h.movieService.GetShuffledMovies()
	if err != nil {
		http.Error(w, "failed to get movies", http.StatusInternalServerError)
		return
	}

	if len(shuffledMovies) > numberOfMovies {
		shuffledMovies = shuffledMovies[:numberOfMovies]
	}

	web.RenderPage(pages.Shuffle(shuffledMovies), "Movies", w, r)
}

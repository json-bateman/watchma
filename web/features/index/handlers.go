package index

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"
	"watchma/db/sqlcgen"
	appctx "watchma/pkg/context"
	"watchma/pkg/jellyfin"
	"watchma/pkg/movie"
	"watchma/web"
	"watchma/web/features/index/pages"
	"watchma/web/views/http_error"
)

type handlers struct {
	movieService *movie.Service
	queries      *sqlcgen.Queries
}

func newHandlers(ms *movie.Service, queries *sqlcgen.Queries) *handlers {
	return &handlers{
		movieService: ms,
		queries:      queries,
	}
}

func (h *handlers) index(w http.ResponseWriter, r *http.Request) {
	movieOfTheDay, err := h.movieService.GetMovieOfTheDay()

	if err != nil {
		// Check if it's an HTTP error from Jellyfin
		var httpErr *jellyfin.HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusUnauthorized {
			web.RenderPage(http_error.JellyfinUnauthorized(), "Jellyfin Unauthorized", w, r)
			return
		}
		http.Error(w, "Failed to load movie of the day", http.StatusInternalServerError)
		return
	}

	web.RenderPage(pages.IndexPage(movieOfTheDay), "Watchma", w, r)
}

// Shuffle returns a page with a shuffled list of movies, up to the number requested in the query parameters
func (h *handlers) shuffle(w http.ResponseWriter, r *http.Request) {
	moviesParam := r.URL.Query().Get("movies")
	numberOfMovies := 9 // default
	if moviesParam != "" {
		n, err := strconv.Atoi(moviesParam)
		if err != nil {
			http.Error(w, "movies param must be a number", http.StatusBadRequest)
			return
		}
		numberOfMovies = n
	}

	shuffledMovies, err := h.movieService.GetShuffledMovies()
	if err != nil {
		http.Error(w, "failed to get movies", http.StatusInternalServerError)
		return
	}

	if len(shuffledMovies) > numberOfMovies {
		shuffledMovies = shuffledMovies[:numberOfMovies]
	}

	web.RenderPage(pages.Shuffle(shuffledMovies, numberOfMovies), "Movies", w, r)
}

// Top5 returns a page with the top5 winning movies across all games played
func (h *handlers) top5(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topWinners, err := h.queries.GetMostPopularWinningMovies(ctx, 5)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "failed to get top winners", http.StatusInternalServerError)
		return
	}

	web.RenderPage(pages.Top5(topWinners), "Top 5 Winners", w, r)
}

func (h *handlers) stats(w http.ResponseWriter, r *http.Request) {
	user := appctx.GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get drafted movies (net count)
	draftedMovies, err := h.queries.GetUserMovieDraftCounts(ctx, user.ID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	// Get voted movies (net count)
	votedMovies, err := h.queries.GetUserMovieVoteCounts(ctx, user.ID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	// Removed topWinners from here - now on separate page
	web.RenderPage(pages.Stats(user, draftedMovies, votedMovies), "Stats", w, r)
}

package web

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"

	"watchma/pkg/types"
	"watchma/pkg/utils"
	"watchma/view/common"
	"watchma/view/movies"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) Shuffle(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")
	user := utils.GetUserFromContext(r)

	intVal, err := strconv.Atoi(number)
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "param must be a number", http.StatusBadRequest)
		return
	}

	_movies, err := h.movieService.GetMovies()
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "Unable to load movies", http.StatusInternalServerError)
		return
	}

	if len(_movies) == 0 {
		log.Printf("no movies found")
		return
	}

	rand.Shuffle(len(_movies), func(i, j int) {
		_movies[i], _movies[j] = _movies[j], _movies[i]
	})

	var randMovies []types.Movie
	if len(_movies) >= intVal {
		randMovies = _movies[:intVal]
	} else {
		randMovies = _movies
	}

	component := movies.Shuffle(common.PageContext{
		Title: "Movies",
		User:  user,
	}, randMovies, h.settings.JellyfinBaseURL)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) SubmitMovies(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	currentUser := utils.GetUserFromContext(r)
	if currentUser == nil {
		utils.SendSSEError(w, r, "Unauthorized")
		return
	}

	var moviesReq types.MovieRequest
	fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&moviesReq); err != nil {
		utils.SendSSEError(w, r, "Invalid Request Body")
		return
	}

	if len(moviesReq.Movies) == 0 {
		utils.SendSSEError(w, r, "Must include at least 1 movie id.")
		return
	}

	h.roomService.SubmitVotes(roomName, currentUser.Username, moviesReq.Movies)
	isVotingFinished := h.roomService.GetIsVotingFinished(roomName)

	sse := datastar.NewSSE(w, r)

	if isVotingFinished {
		h.roomService.FinishGame(roomName)
	} else {
		room, _ := h.roomService.GetRoom(roomName)
		player, _ := room.GetPlayer(currentUser.Username)

		buttonAndMovies := movies.SubmitButton(room.Game.Movies, h.settings.JellyfinBaseURL, player.SelectedMovies)
		sse.PatchElementTempl(buttonAndMovies)
	}
}

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
	"watchma/view/movies"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) Shuffle(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	intVal, err := strconv.Atoi(number)
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "param must be a number", http.StatusBadRequest)
		return
	}

	_movies, err := h.services.MovieService.GetMovies()
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

	response := NewPageResponse(movies.Shuffle(randMovies, h.settings.JellyfinBaseURL), "Movies")
	h.RenderPage(response, w, r)
}

func (h *WebHandler) SubmitMovies(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	currentUser := utils.GetUserFromContext(r)
	if currentUser == nil {
		h.logger.Error("No User found from session cookie")
		return
	}

	var moviesReq types.MovieRequest
	fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&moviesReq); err != nil {
		h.logger.Error("Error decoding movie request", "moviesReq", moviesReq)
		return
	}

	if len(moviesReq.Movies) == 0 {
		utils.SendSSEError(w, r, "Must include at least 1 movie id.")
		return
	}

	h.services.RoomService.SubmitVotes(roomName, currentUser.Username, moviesReq.Movies)
	isVotingFinished := h.services.RoomService.GetIsVotingFinished(roomName)

	sse := datastar.NewSSE(w, r)

	if isVotingFinished {
		h.services.RoomService.FinishGame(roomName)
	} else {
		room, _ := h.services.RoomService.GetRoom(roomName)
		player, _ := room.GetPlayer(currentUser.Username)

		buttonAndMovies := movies.SubmitButton(room.Game.Movies, h.settings.JellyfinBaseURL, player.SelectedMovies)
		sse.PatchElementTempl(buttonAndMovies)
	}
}

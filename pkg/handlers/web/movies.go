package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"watchma/pkg/types"
	"watchma/pkg/utils"
	"watchma/view/shuffle"
	"watchma/view/steps"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
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

	isVotingFinished := h.services.RoomService.SubmitVotes(roomName, currentUser.Username, moviesReq.Movies)

	if isVotingFinished {
		h.services.RoomService.FinishGame(roomName)
	} else {
		room, _ := h.services.RoomService.GetRoom(roomName)
		player, _ := room.GetPlayer(currentUser.Username)

		buttonAndMovies := steps.SubmitButton(room.Game.Movies, h.settings.JellyfinBaseURL, player.SelectedMovies)

		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(buttonAndMovies)
	}
}

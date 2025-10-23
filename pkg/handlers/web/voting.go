package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"watchma/pkg/types"
	"watchma/view/steps"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) SubmitMovies(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	room, ok := h.services.RoomService.GetRoom(roomName)
	if !ok {
		h.logger.Error("Could not obtain room", "room", roomName)
		return
	}

	currentUser := h.GetUserFromContext(r)
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
		h.SendSSEError(w, r, "Must include at least 1 movie id.")
		return
	}

	isVotingFinished := h.services.RoomService.SubmitFinalVotes(roomName, currentUser.Username, moviesReq.Movies)

	// This will advance to Voting, then Results
	if isVotingFinished {
		room.Game.Step += 1
		for _, p := range room.Players {
			p.DraftMovies = []string{}
			p.HasSelectedMovies = false
		}
	} else {
		room, _ := h.services.RoomService.GetRoom(roomName)
		player, _ := room.GetPlayer(currentUser.Username)

		buttonAndMovies := steps.SubmitButton(room.Game.AllMovies, h.settings.JellyfinBaseURL, player.DraftMovies)

		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(buttonAndMovies)
	}
}

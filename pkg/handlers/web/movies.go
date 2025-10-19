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

	items, err := h.movieService.FetchJellyfinMovies()
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "Unable to load movies", http.StatusInternalServerError)
		return
	}

	if items == nil || len(items.Items) == 0 {
		log.Printf("no movies found")
		return
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

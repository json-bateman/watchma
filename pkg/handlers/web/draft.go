package web

import (
	"encoding/json"
	"net/http"
	"watchma/pkg/services"
	"watchma/pkg/types"
	"watchma/view/steps"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

type MovieQueryRequest struct {
	Search string `json:"search"`
	Genre  string `json:"genre"`
	Sort   string `json:"sort"`
}

func (h *WebHandler) DeleteFromSelectedMovies(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	movieId := chi.URLParam(r, "id")
	user := h.GetUserFromContext(r)

	// Use RoomService to handle business logic
	if !h.services.RoomService.RemoveDraftMovie(roomName, user.Username, movieId) {
		h.logger.Error("Failed to remove draft movie", "Room", roomName, "Username", user.Username, "MovieId", movieId)
	}

	h.RenderDraftPage(w, r)
}

func (h *WebHandler) ToggleDraftMovie(w http.ResponseWriter, r *http.Request) {
	movieId := chi.URLParam(r, "id")
	user := h.GetUserFromContext(r)
	roomName := chi.URLParam(r, "roomName")
	room, _ := h.services.RoomService.GetRoom(roomName)

	movie := room.Game.AllMoviesMap[movieId]

	if !h.services.RoomService.ToggleDraftMovie(roomName, user.Username, *movie) {
		h.logger.Error("Failed to toggle draft movie", "Room", roomName, "Username", user.Username, "MovieId", movieId)
	}

	h.RenderDraftPage(w, r)
}

func (h *WebHandler) QueryMovies(w http.ResponseWriter, r *http.Request) {
	h.RenderDraftPage(w, r)
}

func (h *WebHandler) DraftSubmit(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	room, ok := h.services.RoomService.GetRoom(roomName)

	currentUser := h.GetUserFromContext(r)
	if currentUser == nil {
		h.logger.Error("No User found from session cookie")
		return
	}

	player, ok := room.GetPlayer(currentUser.Username)
	if !ok {
		h.logger.Error("Player not in room")
		return
	}
	if len(player.DraftMovies) == 0 {
		h.SendSSEError(w, r, "Must include at least 1 movie id.")
		return
	}

	player.HasFinishedDraft = true

	isDraftFinished := room.IsDraftFinished()

	// if voting is finished, add all players choices to the voting array
	if isDraftFinished {
		room.Game.Step = types.Voting
		h.services.RoomService.SubmitDraftVotes(room)
		h.services.RoomService.MoveToVoting(roomName)
	} else {
		h.RenderDraftPage(w, r)
	}
}

func (h *WebHandler) RenderDraftPage(w http.ResponseWriter, r *http.Request) {
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

	player, ok := room.GetPlayer(currentUser.Username)
	if !ok {
		h.logger.Error("Player not in room")
		return
	}

	var queryRequest MovieQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&queryRequest); err != nil {
		h.logger.Error("Error decoding SortFilter", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var sortField services.MovieSortField
	descending := false
	switch queryRequest.Sort {
	case "name-asc":
		sortField = services.SortByName
	case "name-desc":
		sortField = services.SortByName
		descending = true
	case "year-asc":
		sortField = services.SortByYear
	case "year-desc":
		sortField = services.SortByYear
		descending = true
	case "critic-asc":
		sortField = services.SortByCriticRating
	case "critic-desc":
		sortField = services.SortByCriticRating
		descending = true
	case "community-asc":
		sortField = services.SortByCommunityRating
	case "community-desc":
		sortField = services.SortByCommunityRating
		descending = true
	default:
		// default to name instead of blowing up for now
		// sortField = services.SortByName
	}

	movies, err := h.services.MovieService.GetMoviesWithQuery(
		services.MovieQuery{
			Search:     queryRequest.Search,
			Genre:      queryRequest.Genre,
			SortBy:     sortField,
			Descending: descending,
		},
	)

	if err != nil {
		h.logger.Error("Movie Query Error", "Error", err)
	}
	draft := steps.Draft(player, movies, room)
	if err := datastar.NewSSE(w, r).PatchElementTempl(draft); err != nil {
		h.logger.Error("Error Rendering Draft Page")
	}
}

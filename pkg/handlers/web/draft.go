package web

import (
	"encoding/json"
	"net/http"
	"watchma/pkg/services"
	"watchma/pkg/utils"
	"watchma/view/steps"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

// func (h *WebHandler) JoinDraft(w http.ResponseWriter, r *http.Request) {
// 	movies, _ := h.services.MovieService.GetMovies()
// 	response := NewPageResponse(steps.Draft(testDraftState, movies, h.settings.JellyfinBaseURL), "Draft")
// 	h.RenderPage(response, w, r)
// }

func (h *WebHandler) DeleteFromSelectedMovies(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	movieId := chi.URLParam(r, "id")
	user := utils.GetUserFromContext(r)
	movies, _ := h.services.MovieService.GetMovies()

	myRoom, ok := h.services.RoomService.GetRoom(roomName)
	if ok {
		h.logger.Error("Error finding room", "Room", roomName)
	}

	player, ok := myRoom.GetPlayer(user.Username)
	if !ok {
		h.logger.Error("User not in room", "Username", user.Username, "Room", myRoom.Name)
		return
	}
	// This business logic needs to be put into roomService later
	for i, id := range player.SelectedMovies {
		if id == movieId {
			player.SelectedMovies = append(
				player.SelectedMovies[:i],
				player.SelectedMovies[i+1:]...,
			)
			break
		}
	}

	draftContainerTempl := steps.Draft(
		player,
		movies,
		h.settings.JellyfinBaseURL,
		myRoom,
	)

	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(draftContainerTempl)
}

func (h *WebHandler) ToggleSelectedMovie(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	movieId := chi.URLParam(r, "id")
	movies, _ := h.services.MovieService.GetMovies()
	user := utils.GetUserFromContext(r)

	myRoom, ok := h.services.RoomService.GetRoom(roomName)
	if ok {
		h.logger.Error("Error finding room", "Room", roomName)
	}

	player, ok := myRoom.GetPlayer(user.Username)
	if !ok {
		h.logger.Error("User not in room", "Username", user.Username, "Room", myRoom.Name)
		return
	}

	// This business logic needs to be put into roomService later
	found := false
	for i, id := range player.SelectedMovies {
		if id == movieId {
			player.SelectedMovies = append(
				player.SelectedMovies[:i],
				player.SelectedMovies[i+1:]...,
			)
			found = true
			break
		}
	}

	if !found && len(player.SelectedMovies) < myRoom.Game.MaxDraftCount {
		for _, m := range movies {
			if m.Id == movieId {
				player.SelectedMovies = append(player.SelectedMovies, m.Id)
				break
			}
		}
	}

	draftContainerTempl := steps.Draft(
		player,
		movies,
		h.settings.JellyfinBaseURL,
		myRoom,
	)

	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(draftContainerTempl)
}

func (h *WebHandler) QueryMovies(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := utils.GetUserFromContext(r)
	movies, _ := h.services.MovieService.GetMovies()

	myRoom, ok := h.services.RoomService.GetRoom(roomName)
	if ok {
		h.logger.Error("Error finding room", "Room", roomName)
	}

	player, ok := myRoom.GetPlayer(user.Username)
	if !ok {
		h.logger.Error("User not in room", "Username", user.Username, "Room", myRoom.Name)
		return
	}

	type MovieQueryRequest struct {
		Search string `json:"search"`
		Genre  string `json:"genre"`
		Sort   string `json:"sort"`
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
		sortField = services.SortByName
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
		http.Error(w, "Failed to get movies", http.StatusInternalServerError)
		return
	}

	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(steps.Draft(
		player,
		movies,
		h.settings.JellyfinBaseURL,
		myRoom,
	))
}

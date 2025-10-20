package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"watchma/pkg/types"
	"watchma/view/draft"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

var testDraftState = types.DraftState{
	MaxVotes: 8,
	SelectedMovies: []types.Movie{
		{
			CommunityRating: 7.545,
			CriticRating:    88,
			Genres:          []string{"Family", "Comedy", "Crime", "Adventure", "Animation"},
			Id:              "eeabfd0d5436e34e85fe977afc1c54d5",
			Name:            "The Bad Guys",
			PremiereDate:    "2022-03-17T00:00:00.0000000Z",
			PrimaryImageTag: "d377231308c50295fa54c052a00836d1",
			ProductionYear:  2022,
		},
		{
			CommunityRating: 7.918,
			CriticRating:    84,
			Genres:          []string{"Action", "Thriller"},
			Id:              "202b05b82b0a1b8eb0e42d785c981bd7",
			Name:            "Nobody",
			PremiereDate:    "2021-03-18T00:00:00.0000000Z",
			PrimaryImageTag: "fce50f2a76aa55feec713720063e87b5",
			ProductionYear:  2021,
		},
	},
	IsReady: false,
}

func (h *WebHandler) JoinDraft(w http.ResponseWriter, r *http.Request) {
	movies, _ := h.services.MovieService.GetMovies()
	response := NewPageResponse(draft.Draft(testDraftState, movies, h.settings.JellyfinBaseURL), "Draft")
	h.RenderPage(response, w, r)
}

func (h *WebHandler) DeleteFromSelectedMovies(w http.ResponseWriter, r *http.Request) {
	movieId := chi.URLParam(r, "id")
	movies, _ := h.services.MovieService.GetMovies()
	sse := datastar.NewSSE(w, r)

	// This business logic needs to be put into roomService later
	for i, m := range testDraftState.SelectedMovies {
		if m.Id == movieId {
			testDraftState.SelectedMovies = append(
				testDraftState.SelectedMovies[:i],
				testDraftState.SelectedMovies[i+1:]...,
			)
			break
		}
	}

	// now render with the updated testDraftState
	draftContainerTempl := draft.Draft(
		testDraftState,
		movies,
		h.settings.JellyfinBaseURL,
	)

	_ = sse.PatchElementTempl(draftContainerTempl)
}

func (h *WebHandler) ToggleSelectedMovie(w http.ResponseWriter, r *http.Request) {
	movieId := chi.URLParam(r, "id")
	movies, _ := h.services.MovieService.GetMovies()
	sse := datastar.NewSSE(w, r)

	// This business logic needs to be put into roomService later
	found := false
	for i, m := range testDraftState.SelectedMovies {
		if m.Id == movieId {
			testDraftState.SelectedMovies = append(
				testDraftState.SelectedMovies[:i],
				testDraftState.SelectedMovies[i+1:]...,
			)
			found = true
			break
		}
	}

	// if not found, add it
	if !found {
		for _, m := range movies {
			if m.Id == movieId {
				testDraftState.SelectedMovies = append(testDraftState.SelectedMovies, m)
				break
			}
		}
	}

	// Send new patch to frontend
	draftContainerTempl := draft.Draft(
		testDraftState,
		movies,
		h.settings.JellyfinBaseURL,
	)
	_ = sse.PatchElementTempl(draftContainerTempl)
}

type SortFilter struct {
	Search string `json:"search"`
	Genre  string `json:"genre"`
	Sort   string `json:"sort"`
}

func (h *WebHandler) SortAndFilterMovies(w http.ResponseWriter, r *http.Request) {

	var sortFilter SortFilter
	if err := json.NewDecoder(r.Body).Decode(&sortFilter); err != nil {
		h.logger.Error("Error decoding SortFilter", "SortFilter", sortFilter)
		return
	}
	fmt.Printf("%+v", sortFilter)
}

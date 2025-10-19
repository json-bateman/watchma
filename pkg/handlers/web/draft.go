package web

import (
	"net/http"
	"watchma/pkg/types"
	"watchma/pkg/utils"
	"watchma/view/common"
	"watchma/view/draft"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

var testDraftState = types.DraftState{
	MaxVotes: 8,
	SelectedMovies: []types.JellyfinItem{
		{
			Name:            "The Bad Guys",
			Id:              "eeabfd0d5436e34e85fe977afc1c54d5",
			Container:       "mkv",
			PremiereDate:    "2022-03-17T00:00:00.0000000Z",
			CriticRating:    88,
			CommunityRating: 7.545,
			RunTimeTicks:    60070060000,
			ProductionYear:  2022,
			ImageTags: struct {
				Primary string `json:"Primary"`
				Logo    string `json:"Logo"`
				Thumb   string `json:"Thumb"`
			}{
				Primary: "d377231308c50295fa54c052a00836d1",
				Logo:    "ae7f44f6d988e3c72f4a94369ba44ba1",
				Thumb:   "a91298fcff4e668ee0fbed0b6f4088ba",
			},
			BackdropImageTags: []string{
				"02f175b9bd50e77ed8f32d87e18de4b3",
			},
			Genres: []string{
				"Family",
				"Comedy",
				"Crime",
				"Adventure",
				"Animation",
			},
		},
		{
			Name:            "Nobody",
			Id:              "202b05b82b0a1b8eb0e42d785c981bd7",
			Container:       "mkv",
			PremiereDate:    "2021-03-18T00:00:00.0000000Z",
			CriticRating:    84,
			CommunityRating: 7.918,
			RunTimeTicks:    55018660000,
			ProductionYear:  2021,
			ImageTags: struct {
				Primary string `json:"Primary"`
				Logo    string `json:"Logo"`
				Thumb   string `json:"Thumb"`
			}{
				Primary: "fce50f2a76aa55feec713720063e87b5",
				Logo:    "c85444a25557bded381a8f2907c0d03d",
				Thumb:   "0ae30a15d851a0e1d87cb157056b3647",
			},
			BackdropImageTags: []string{
				"492bb1513d9ff4f1e6b9521b1b5e94e0",
			},
			Genres: []string{
				"Action",
				"Thriller",
			},
		},
	},
	IsReady: false,
}

func (h *WebHandler) JoinDraft(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r)
	movies, _ := h.movieService.FetchJellyfinMovies()

	templ.WithChildren(r.Context(), draft.Container(testDraftState, movies.Items, h.settings.JellyfinBaseURL))

	templ.Handler(draft.Draft(
		common.PageContext{
			Title: "Draft",
			User:  user,
		},
		testDraftState,
		movies.Items, h.settings.JellyfinBaseURL)).ServeHTTP(w, r)
}

func (h *WebHandler) DeleteFromSelectedMovies(w http.ResponseWriter, r *http.Request) {
	movieId := chi.URLParam(r, "id")
	movies, _ := h.movieService.FetchJellyfinMovies()
	sse := datastar.NewSSE(w, r)

	// put this in the room service
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
	draftContainerTempl := draft.Container(
		testDraftState,
		movies.Items,
		h.settings.JellyfinBaseURL,
	)

	_ = sse.PatchElementTempl(draftContainerTempl)
}

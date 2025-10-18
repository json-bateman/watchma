package web

import (
	"net/http"
	"watchma/pkg/types"
	"watchma/view/draft"

	"github.com/a-h/templ"
)

func (h *WebHandler) JoinDraft(w http.ResponseWriter, r *http.Request) {
	movies, _ := h.movieService.FetchJellyfinMovies()

	templ.Handler(draft.Draft(types.DraftState{
		MaxVotes: 8,
		SelectedMovies: []types.JellyfinItem{
			{
				Name:            "The Matrix",
				Id:              "movie-1",
				Container:       "mkv",
				PremiereDate:    "1999-03-31T00:00:00Z",
				CriticRating:    88,
				CommunityRating: 8.7,
				RunTimeTicks:    81600000000, // 2h 16m in ticks
				ProductionYear:  1999,
				ImageTags: struct {
					Primary string `json:"Primary"`
					Logo    string `json:"Logo"`
					Thumb   string `json:"Thumb"`
				}{
					Primary: "matrix-poster",
					Logo:    "matrix-logo",
					Thumb:   "matrix-thumb",
				},
				BackdropImageTags: []string{"matrix-backdrop"},
			},
			{
				Name:            "Inception",
				Id:              "movie-2",
				Container:       "mkv",
				PremiereDate:    "2010-07-16T00:00:00Z",
				CriticRating:    87,
				CommunityRating: 8.8,
				RunTimeTicks:    88800000000, // 2h 28m in ticks
				ProductionYear:  2010,
				ImageTags: struct {
					Primary string `json:"Primary"`
					Logo    string `json:"Logo"`
					Thumb   string `json:"Thumb"`
				}{
					Primary: "inception-poster",
					Logo:    "inception-logo",
					Thumb:   "inception-thumb",
				},
				BackdropImageTags: []string{"inception-backdrop"},
			}},
		IsReady: false,
	}, movies.Items, h.settings.JellyfinBaseURL)).ServeHTTP(w, r)
}

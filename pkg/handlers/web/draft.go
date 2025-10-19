package web

import (
	"net/http"
	"watchma/pkg/types"
	"watchma/pkg/utils"
	"watchma/view/common"
	"watchma/view/draft"

	"github.com/a-h/templ"
)

func (h *WebHandler) JoinDraft(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r)
	movies, _ := h.movieService.GetMovies()

	templ.Handler(draft.Draft(common.PageContext{
		Title: "Draft",
		User:  user,
	}, types.DraftState{
		MaxVotes:       8,
		SelectedMovies: []types.Movie{},
		IsReady:        false,
	}, movies, h.settings.JellyfinBaseURL)).ServeHTTP(w, r)
}

package web

import (
	"net/http"
	"watchma/view/draft"

	"github.com/a-h/templ"
)

func (h *WebHandler) JoinDraft(w http.ResponseWriter, r *http.Request) {
	templ.Handler(draft.Draft()).ServeHTTP(w, r)
}

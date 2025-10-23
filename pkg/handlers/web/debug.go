package web

import (
	"net/http"
	"watchma/view/debug"

	"github.com/a-h/templ"
)

func (h *WebHandler) Debug(w http.ResponseWriter, r *http.Request) {
	debugSnapshot := h.services.RoomService.GetDebugSnapshot()

	component := debug.Debug(debugSnapshot)
	templ.Handler(component).ServeHTTP(w, r)
}

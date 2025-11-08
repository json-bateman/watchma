package debug

import (
	"net/http"
	"watchma/features/debug/pages"
	appctx "watchma/pkg/context"
	"watchma/pkg/services"

	"github.com/a-h/templ"
)

type handlers struct {
	roomService *services.RoomService
}

func newHandlers(rs *services.RoomService) *handlers {
	return &handlers{
		roomService: rs,
	}
}

// renderPage is a shared helper to render pages with the common layout
func (h *handlers) debug(w http.ResponseWriter, r *http.Request) {
	user := appctx.GetUserFromRequest(r)

	debugSnapshot := h.roomService.GetDebugSnapshot()

	component := pages.Debug(debugSnapshot, user)
	templ.Handler(component).ServeHTTP(w, r)
}

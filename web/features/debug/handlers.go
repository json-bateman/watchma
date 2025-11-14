package debug

import (
	"log/slog"
	"net/http"
	appctx "watchma/pkg/context"
	"watchma/pkg/room"
	"watchma/web"
	"watchma/web/features/debug/pages"

	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"
)

type handlers struct {
	roomService *room.Service
	logger      *slog.Logger
	nats        *nats.Conn
}

func newHandlers(rs *room.Service, logger *slog.Logger, nc *nats.Conn) *handlers {
	return &handlers{
		roomService: rs,
		logger:      logger,
		nats:        nc,
	}
}

// renderPage is a shared helper to render pages with the common layout
func (h *handlers) debug(w http.ResponseWriter, r *http.Request) {
	user := appctx.GetUserFromRequest(r)

	debugSnapshot := h.roomService.GetDebugSnapshot()

	web.RenderPageNoLayout(pages.Debug(debugSnapshot, user), "debug", w, r)
}

func (h *handlers) sse(w http.ResponseWriter, r *http.Request) {
	user := appctx.GetUserFromRequest(r)
	sse := datastar.NewSSE(w, r)

	// Send initial debug snapshot to new client
	debugSnapshot := h.roomService.GetDebugSnapshot()
	if err := sse.PatchElementTempl(pages.Debug(debugSnapshot, user)); err != nil {
		h.logger.Error("Error patching initial debug snapshot")
		return
	}

	// Subscribe to all NATS subjects using wildcard
	sub, err := h.nats.SubscribeSync(">")
	h.logger.Debug("NATS Subscribe", "subject", ">", "description", "all events")
	defer sub.Unsubscribe()
	if err != nil {
		http.Error(w, "Subscribe Failed", http.StatusInternalServerError)
		return
	}

	// Listen for any NATS events and update debug page
	for {
		msg, err := sub.NextMsgWithContext(r.Context())
		if err != nil {
			// context canceled or sub closed
			return
		}

		// Log the event for debugging
		h.logger.Debug("NATS Event Received", "subject", msg.Subject, "data", string(msg.Data))

		// Refresh debug snapshot on any event
		debugSnapshot := h.roomService.GetDebugSnapshot()
		if err := sse.PatchElementTempl(pages.Debug(debugSnapshot, user)); err != nil {
			h.logger.Error("Error patching debug snapshot", "error", err)
			return
		}
	}
}

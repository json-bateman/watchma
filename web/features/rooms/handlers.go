package rooms

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	appctx "watchma/pkg/context"
	"watchma/pkg/movie"
	"watchma/pkg/room"
	"watchma/web"
	"watchma/web/features/rooms/pages"

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

func (h *handlers) join(w http.ResponseWriter, r *http.Request) {
	web.RenderPage(pages.JoinPage(h.roomService.Rooms), "Join Room", w, r)
}

func (h *handlers) joinSSE(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Send initial room list to new client
	roomList := pages.RoomListBody(h.roomService.Rooms)
	if err := sse.PatchElementTempl(roomList); err != nil {
		h.logger.Error("Error patching initial room list", "error", err)
	}

	sub, err := h.nats.SubscribeSync(room.NATSLobbyRooms)
	h.logger.Debug(room.NATSSub, "subject", room.NATSLobbyRooms)
	defer sub.Unsubscribe()
	if err != nil {
		http.Error(w, "Subscribe Failed", http.StatusInternalServerError)
		return
	}

	for {
		msg, err := sub.NextMsgWithContext(r.Context())
		if err != nil {
			// context canceled or sub closed
			return
		}
		switch string(msg.Data) {
		case room.RoomListUpdateEvent:
			roomList := pages.RoomListBody(h.roomService.Rooms)
			if err := sse.PatchElementTempl(roomList); err != nil {
				h.logger.Error("Error patching room list", "error", err)
				return
			}
		default: // discard unknown non-matching messages
		}
	}
}

func (h *handlers) host(w http.ResponseWriter, r *http.Request) {
	web.RenderPage(pages.HostPage(), "Host Room", w, r)
}

func (h *handlers) hostForm(w http.ResponseWriter, r *http.Request) {
	user := appctx.GetUserFromRequest(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		h.logger.Error("User was nil")
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		h.logger.Warn("Error parsing form", "error", err)
		return
	}

	roomName := r.FormValue("roomName")

	movies, err := atoiField(r, "draftNumber")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	maxPlayers, err := atoiField(r, "maxplayers")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.roomService.RoomExists(roomName) {
		http.Error(w, "This room name already exists", http.StatusConflict)
		return
	}
	h.roomService.AddRoom(roomName, &room.Session{
		MaxDraftCount: movies,
		MaxPlayers:    maxPlayers,
		Host:          user.Username,
		Votes:         make(map[*movie.Movie]int),
	})

	http.Redirect(w, r, fmt.Sprintf("/room/%s/lobby", roomName), http.StatusSeeOther)
}

func atoiField(r *http.Request, key string) (int, error) {
	v := r.FormValue(key)
	if v == "" {
		return 0, fmt.Errorf("missing field %s", key)
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return i, nil
}

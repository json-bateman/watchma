package rooms

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"watchma/features/rooms/pages"
	appctx "watchma/pkg/context"
	"watchma/pkg/services"
	"watchma/pkg/types"
	"watchma/pkg/web"

	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"
)

type handlers struct {
	roomService *services.RoomService
	logger      *slog.Logger
	nats        *nats.Conn
}

func newHandlers(rs *services.RoomService, logger *slog.Logger, nc *nats.Conn) *handlers {
	return &handlers{
		roomService: rs,
		logger:      logger,
		nats:        nc,
	}
}

func (h *handlers) join(w http.ResponseWriter, r *http.Request) {
	web.RenderPage(pages.JoinPage(h.roomService.Rooms), "Join page", w, r)
}

func (h *handlers) joinSSE(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Send initial room list to new client
	roomList := pages.RoomListBody(h.roomService.Rooms)
	if err := sse.PatchElementTempl(roomList); err != nil {
		h.logger.Error("Error patching initial room list")
	}

	sub, err := h.nats.SubscribeSync(types.NATS_LOBBY_ROOMS)
	h.logger.Debug(types.NATS_SUB, "subject", types.NATS_LOBBY_ROOMS)
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
		case types.ROOM_LIST_UPDATE_EVENT:
			roomList := pages.RoomListBody(h.roomService.Rooms)
			if err := sse.PatchElementTempl(roomList); err != nil {
				fmt.Println("Error patching room list")
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
		h.logger.Error("Error parsing form", "error", err)
		return
	}

	roomName := r.FormValue("roomName")
	moviesStr := r.FormValue("draftNumber")
	maxPlayersStr := r.FormValue("maxplayers")
	maxVotesStr := r.FormValue("maxvotes")
	displayTies := r.FormValue("displayTies")

	movies, err := strconv.Atoi(moviesStr)
	maxPlayers, err := strconv.Atoi(maxPlayersStr)
	maxVotes, err := strconv.Atoi(maxVotesStr)
	if err != nil {
		http.Error(w, "Error converting form strings", http.StatusInternalServerError)
		h.logger.Error("Error converting form strings", "error", err)
		return
	}
	if h.roomService.RoomExists(roomName) {
		http.Error(w, "This room name already exists", http.StatusConflict)
		return
	}
	h.roomService.AddRoom(roomName, &types.GameSession{
		MaxDraftCount: movies,
		MaxVotes:      maxVotes,
		MaxPlayers:    maxPlayers,
		DisplayTies:   displayTies == "yes",
		Host:          user.Username,
		Votes:         make(map[*types.Movie]int),
	})

	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusSeeOther)
}

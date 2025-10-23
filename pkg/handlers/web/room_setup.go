package web

import (
	"fmt"
	"net/http"
	"strconv"

	"watchma/pkg/types"
	"watchma/pkg/utils"
	"watchma/view/rooms"

	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) Join(w http.ResponseWriter, r *http.Request) {
	response := NewPageResponse(rooms.JoinPage(h.services.RoomService.Rooms), "Join page")
	h.RenderPage(response, w, r)
}

func (h *WebHandler) JoinSSE(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Send initial room list to new client
	roomList := rooms.RoomListBody(h.services.RoomService.Rooms)
	if err := sse.PatchElementTempl(roomList); err != nil {
		h.logger.Error("Error patching initial room list")
	}

	sub, err := h.NATS.SubscribeSync(utils.NATS_LOBBY_ROOMS)
	h.logger.Debug(utils.NATS_SUB, "subject", utils.NATS_LOBBY_ROOMS)
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
		case utils.ROOM_LIST_UPDATE_EVENT:
			roomList := rooms.RoomListBody(h.services.RoomService.Rooms)
			if err := sse.PatchElementTempl(roomList); err != nil {
				fmt.Println("Error patching room list")
				return
			}
		default: // discard unknown non-matching messages
		}
	}
}

func (h *WebHandler) Host(w http.ResponseWriter, r *http.Request) {
	response := NewPageResponse(rooms.HostPage(), "Host Room")
	h.RenderPage(response, w, r)
}

func (h *WebHandler) HostForm(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	roomName := r.FormValue("roomName")
	moviesStr := r.FormValue("movies")
	maxPlayersStr := r.FormValue("maxplayers")

	movies, err := strconv.Atoi(moviesStr)
	maxPlayers, err := strconv.Atoi(maxPlayersStr)
	if err != nil {
		http.Error(w, "Movies must be a number", http.StatusBadRequest)
		return
	}
	if h.services.RoomService.RoomExists(roomName) {
		http.Error(w, "This room name already exists", http.StatusBadRequest)
		return
	}
	h.services.RoomService.AddRoom(roomName, &types.GameSession{
		MaxDraftCount: movies,
		MaxPlayers:    maxPlayers,
		Host:          user.Username,
		Votes:         make(map[*types.Movie]int),
	})

	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusSeeOther)
}

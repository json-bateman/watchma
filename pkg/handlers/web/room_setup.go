package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
	"watchma/pkg/types"
	"watchma/pkg/utils"
	"watchma/view/rooms"
)

func (h *WebHandler) Join(w http.ResponseWriter, r *http.Request) {
	component := rooms.JoinPage(h.roomService.Rooms)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) JoinSSE(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Send initial room list to new client
	roomList := rooms.RoomListBody(h.roomService.Rooms)
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
			roomList := rooms.RoomListBody(h.roomService.Rooms)
			if err := sse.PatchElementTempl(roomList); err != nil {
				fmt.Println("Error patching room list")
				return
			}
		default: // discard unknown non-matching messages
		}
	}
}

func (h *WebHandler) Host(w http.ResponseWriter, r *http.Request) {
	component := rooms.HostPage("username")
	templ.Handler(component).ServeHTTP(w, r)
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
	if h.roomService.RoomExists(roomName) {
		http.Error(w, "This room name already exists", http.StatusBadRequest)
		return
	}
	h.roomService.AddRoom(roomName, &types.GameSession{
		MovieNumber: movies,
		MaxPlayers:  maxPlayers,
		Host:        user.Username,
		Votes:       make(map[*types.JellyfinItem]int),
	})

	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusSeeOther)
}

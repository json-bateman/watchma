package web

import (
	"net/http"

	"watchma/pkg/types"
	"watchma/view/rooms"
	"watchma/view/steps"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) SingleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := h.GetUserFromContext(r)

	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	myRoom, ok := h.services.RoomService.GetRoom(roomName)

	if !ok {
		h.RenderPage(rooms.NoRoom(roomName), roomName, w, r)
		return
	}

	if myRoom.Game.MaxPlayers <= len(myRoom.Players) {
		h.RenderPage(rooms.RoomFull(), roomName, w, r)
		return
	}

	h.services.RoomService.AddPlayerToRoom(myRoom.Name, user.Username)

	h.RenderPageNoLayout(steps.Lobby(myRoom, user.Username), myRoom.Name, w, r)
}

// Function that does the heavy lifting by keeping the SSE channel open and sending
// Events to the client in real-time
func (h *WebHandler) SingleRoomSSE(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := h.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sse := datastar.NewSSE(w, r)

	// Check if room exists
	myRoom, ok := h.services.RoomService.GetRoom(roomName)
	if !ok {
		h.logger.Error("Room not found on SSE reconnect", "Room", roomName, "Username", user.Username)
		// Send error and redirect to home
		if err := sse.ExecuteScript("window.location.href = '/'"); err != nil {
			h.logger.Error("Error redirecting after room not found", "error", err)
		}
		return
	}

	// Send existing user list to new client
	userBox := steps.UserBox(myRoom, user.Username)
	if err := sse.PatchElementTempl(userBox); err != nil {
		h.logger.Error("Error patching initial user list")
	}

	player, ok := myRoom.GetPlayer(user.Username)
	if !ok {
		h.logger.Error("User not in room", "Username", user.Username, "Room", myRoom.Name)
		// Send error and redirect to home
		if err := sse.ExecuteScript("window.location.href = '/'"); err != nil {
			h.logger.Error("Error redirecting after player not found", "error", err)
		}
		return
	}

	// Send existing messages to new client
	if len(myRoom.RoomMessages) > 0 {
		chat := steps.ChatBox(myRoom.RoomMessages)
		if err := sse.PatchElementTempl(chat); err != nil {
			h.logger.Error("Error patching chatbox on load")
			return
		}
	}

	// Subscribe to room-specific NATS subject
	roomSubject := types.RoomSubject(roomName)
	sub, err := h.NATS.SubscribeSync(roomSubject)
	h.logger.Debug(types.NATS_SUB, "subject", roomSubject)
	if err != nil {
		http.Error(w, "Subscribe Failed", http.StatusInternalServerError)
		return
	}

	movies, err := h.services.MovieService.GetMovies()
	if err != nil {
		h.logger.Error("Call to MovieService.GetMovies failed", "Error", err)
		return
	}

	// Cleanup when connection closes
	defer func() {
		sub.Unsubscribe()
		h.LeaveRoom(w, r)
	}()

	for {
		msg, err := sub.NextMsgWithContext(r.Context())
		if err != nil {
			// context canceled or sub closed
			return
		}
		switch string(msg.Data) {
		case types.ROOM_UPDATE_EVENT:
			userBox := steps.UserBox(myRoom, user.Username)
			if err := sse.PatchElementTempl(userBox); err != nil {
				h.logger.Error("Error patching user list", "error", err)
				return
			}
		case types.MESSAGE_SENT_EVENT:
			chat := steps.ChatBox(myRoom.RoomMessages)
			if err := sse.PatchElementTempl(chat); err != nil {
				h.logger.Error("Error patching chat message", "error", err)
				return
			}
		case types.ROOM_START_EVENT:
			draft := steps.Draft(player, movies, myRoom)
			if err := sse.PatchElementTempl(draft); err != nil {
				h.logger.Error("Error patching movies", "error", err)
				return
			}
		case types.ROOM_VOTING_EVENT:
			movies := steps.Voting(myRoom.Game.VotingMovies, player, myRoom)
			if err := sse.PatchElementTempl(movies); err != nil {
				h.logger.Error("Error patching movies", "error", err)
				return
			}
		case types.ROOM_ANNOUNCE_EVENT:
			movies := steps.AiAnnounce(myRoom, []string{""})
			if err := sse.PatchElementTempl(movies); err != nil {
				h.logger.Error("Error patching movies", "error", err)
				return
			}
		case types.ROOM_FINISH_EVENT:
			movieVotes := SortMoviesByVotes(myRoom.Game.Votes)
			winnerMovies := GetWinnerMovies(movieVotes, myRoom)
			finalScreen := steps.ResultsScreen(winnerMovies)
			if err := sse.PatchElementTempl(finalScreen); err != nil {
				h.logger.Error("Error patching final screen", "error", err)
				return
			}

		default: // discard for now, maybe error?
		}
	}
}

func (h *WebHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := h.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := h.services.RoomService.GetRoom(roomName)
	if !ok {
		// Room doesn't exist
		h.WriteJSONError(w, http.StatusNotFound, "Room not found")
		return
	}

	h.services.RoomService.RemovePlayerFromRoom(room.Name, user.Username)

	allUsers := room.GetAllPlayers()
	if len(allUsers) == 0 {
		h.services.RoomService.DeleteRoom(room.Name)
		return
	}

	if room.Game.Host == user.Username {
		// If host leaves transfer to random other user
		for newHostUsername := range room.Players {
			h.services.RoomService.TransferHost(room.Name, newHostUsername)
			break
		}
	}
}

package web

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/json-bateman/jellyfin-grabber/internal/utils"
	"github.com/json-bateman/jellyfin-grabber/view/movies"
	"github.com/json-bateman/jellyfin-grabber/view/rooms"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) SingleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	username := utils.GetUsernameFromCookie(r)

	myRoom, ok := h.roomService.GetRoom(roomName)
	if ok && myRoom.Game.MaxPlayers > len(myRoom.Users) {
		myRoom.AddUser(username)
		h.BroadcastToRoom(roomName, utils.ROOM_UPDATE_EVENT)
		// Broadcast room list update to join page clients
		h.BroadcastToJoinClients(utils.ROOM_LIST_UPDATE_EVENT)
	} else {
		component := rooms.RoomFull(roomName)
		templ.Handler(component).ServeHTTP(w, r)
	}

	if !ok {
		component := rooms.NoRoom(roomName)
		templ.Handler(component).ServeHTTP(w, r)
		return
	}

	component := rooms.SingleRoom(myRoom, username)
	templ.Handler(component).ServeHTTP(w, r)
}

// Function that does the heavy lifting by keeping the SSE channel open and sending
// Events to the client in real-time
func (h *WebHandler) SingleRoomSSE(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	username := utils.GetUsernameFromCookie(r)
	sse := datastar.NewSSE(w, r)

	// Send existing user list to new client
	myRoom, ok := h.roomService.GetRoom(roomName)
	if ok {
		userBox := rooms.UserBox(myRoom, username)
		if err := sse.PatchElementTempl(userBox); err != nil {
			h.logger.Error("Error patching initial user list")
		}
	}

	// Send existing message history to new client
	if ok {
		if len(myRoom.RoomMessages) > 0 {
			chat := rooms.ChatBox(myRoom.RoomMessages)
			if err := sse.PatchElementTempl(chat); err != nil {
				h.logger.Error("Error patching chatbox on load")
				return
			}
		}
	}

	// Subscribe to room-specific NATS subject
	roomSubject := utils.RoomSubject(roomName)
	sub, err := h.NATS.SubscribeSync(roomSubject)
	if err != nil {
		http.Error(w, "Subscribe Failed", http.StatusInternalServerError)
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
		case utils.ROOM_UPDATE_EVENT:
			userBox := rooms.UserBox(myRoom, username)
			if err := sse.PatchElementTempl(userBox); err != nil {
				h.logger.Error("Error patching user list", "error", err)
				return
			}
		case utils.MESSAGE_SENT_EVENT:
			chat := rooms.ChatBox(myRoom.RoomMessages)
			if err := sse.PatchElementTempl(chat); err != nil {
				h.logger.Error("Error patching chat message", "error", err)
				return
			}
		case utils.ROOM_START_EVENT:
			movies := movies.Movies(myRoom.Game.Movies, h.settings.JellyfinBaseURL, myRoom)
			if err := sse.PatchElementTempl(movies); err != nil {
				h.logger.Error("Error patching movies", "error", err)
				return
			}
		case utils.ROOM_FINISH_EVENT:
			// Extract map entries into a slice
			var movieVotes []types.MovieVote
			for jfinMovie, votes := range myRoom.Game.Votes {
				movieVotes = append(movieVotes, types.MovieVote{
					Movie: jfinMovie,
					Votes: votes,
				})
			}

			// Sort by votes (descending - highest votes first)
			sort.Slice(movieVotes, func(i, j int) bool {
				return movieVotes[i].Votes > movieVotes[j].Votes
			})
			finalScreen := movies.GameFinished(movieVotes, h.settings.JellyfinBaseURL)
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
	username := utils.GetUsernameFromCookie(r)

	room, ok := h.roomService.GetRoom(roomName)
	if !ok {
		// Room doesn't exist
		utils.WriteJSONError(w, http.StatusNotFound, "Room not found")
		return
	}

	room.RemoveUser(username)
	allUsers := room.GetAllUsers()
	if len(allUsers) == 0 {
		h.roomService.DeleteRoom(roomName)
		h.BroadcastToJoinClients(utils.ROOM_LIST_UPDATE_EVENT)
		utils.WriteJSONResponse(w, http.StatusOK, map[string]any{
			"ok":      true,
			"room":    roomName,
			"event":   utils.ROOM_UPDATE_EVENT,
			"message": "Room deleted - last user left",
		})
		return
	}

	if room.Game.Host == username {
		// If host leaves transfer to random other user
		for userName := range room.Users {
			room.Game.Host = userName
			break
		}
	}

	h.BroadcastToRoom(roomName, utils.ROOM_UPDATE_EVENT)
	// Broadcast room list update to join page clients
	h.BroadcastToJoinClients(utils.ROOM_LIST_UPDATE_EVENT)

	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{
		"ok":    true,
		"room":  room.Name,
		"event": utils.ROOM_UPDATE_EVENT,
	})
}

func (h *WebHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	room, ok := h.roomService.GetRoom(roomName)
	if ok {
		room.Game.Step = types.Voting
		items, err := h.movieService.FetchJellyfinMovies()
		if err != nil {
		}

		if items == nil || len(items.Items) == 0 {
			h.logger.Info(fmt.Sprintf("Room %s: No Movies Found", room.Name))
		}

		rand.Shuffle(len(items.Items), func(i, j int) {
			items.Items[i], items.Items[j] = items.Items[j], items.Items[i]
		})

		var randMovies []types.JellyfinItem
		if len(items.Items) >= room.Game.MovieNumber {
			randMovies = items.Items[:room.Game.MovieNumber]
		} else {
			randMovies = items.Items
		}
		room.Game.Movies = randMovies

		h.BroadcastToRoom(roomName, utils.ROOM_START_EVENT)
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{
		"ok":    true,
		"room":  room.Name,
		"event": utils.ROOM_START_EVENT,
	})
}

func (h *WebHandler) Ready(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	username := utils.GetUsernameFromCookie(r)

	room, ok := h.roomService.GetRoom(roomName)
	if ok {
		u, found := room.GetUser(username)
		if found && u.Ready {
			u.Ready = false
		} else {
			u.Ready = true
		}
		h.BroadcastToRoom(roomName, utils.ROOM_UPDATE_EVENT)
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{
		"ok":    true,
		"room":  room.Name,
		"event": utils.ROOM_UPDATE_EVENT,
	})
}

func (h *WebHandler) PublishChatMessage(w http.ResponseWriter, r *http.Request) {
	var req types.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}
	if len(strings.Trim(req.Message, " ")) == 0 {
		return
	}

	username := utils.GetUsernameFromCookie(r)
	if username == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, "Username not found")
		return
	}

	req.Username = username
	room, ok := h.roomService.GetRoom(req.Room)
	if ok {
		room.RoomMessages = append(room.RoomMessages, req)
		h.BroadcastToRoom(room.Name, utils.MESSAGE_SENT_EVENT)
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{
		"ok":    true,
		"room":  req.Room,
		"event": utils.MESSAGE_SENT_EVENT,
	})
}

func (h *WebHandler) BroadcastToRoom(roomName, message string) {
	subject := utils.RoomSubject(roomName)
	if err := h.NatsPublish(subject, []byte(message)); err != nil {
		h.logger.Error("Failed to broadcast to room", "room", roomName, "error", err)
	}
}

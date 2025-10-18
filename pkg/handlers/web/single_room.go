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
	"github.com/starfederation/datastar-go/datastar"
	"watchma/pkg/types"
	"watchma/pkg/utils"
	"watchma/view/movies"
	"watchma/view/rooms"
)

func (h *WebHandler) SingleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := utils.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	myRoom, ok := h.roomService.GetRoom(roomName)
	if ok && myRoom.Game.MaxPlayers > len(myRoom.Users) {
		h.roomService.AddUserToRoom(myRoom.Name, user.Username)
	} else {
		component := rooms.RoomFull(roomName)
		templ.Handler(component).ServeHTTP(w, r)
	}

	if !ok {
		component := rooms.NoRoom(roomName)
		templ.Handler(component).ServeHTTP(w, r)
		return
	}

	component := rooms.SingleRoom(myRoom, user.Username)
	templ.Handler(component).ServeHTTP(w, r)
}

// Function that does the heavy lifting by keeping the SSE channel open and sending
// Events to the client in real-time
func (h *WebHandler) SingleRoomSSE(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := utils.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sse := datastar.NewSSE(w, r)

	// Send existing user list to new client
	myRoom, ok := h.roomService.GetRoom(roomName)
	if ok {
		userBox := rooms.UserBox(myRoom, user.Username)
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
	h.logger.Debug(utils.NATS_SUB, "subject", roomSubject)
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
			userBox := rooms.UserBox(myRoom, user.Username)
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
	user := utils.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := h.roomService.GetRoom(roomName)
	if !ok {
		// Room doesn't exist
		utils.WriteJSONError(w, http.StatusNotFound, "Room not found")
		return
	}

	h.roomService.RemoveUserFromRoom(room.Name, user.Username)

	allUsers := room.GetAllUsers()
	if len(allUsers) == 0 {
		h.roomService.DeleteRoom(room.Name)
		return
	}

	if room.Game.Host == user.Username {
		// If host leaves transfer to random other user
		for newHostUsername := range room.Users {
			h.roomService.TransferHost(room.Name, newHostUsername)
			break
		}
	}
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

		h.roomService.StartGame(roomName, room.Game.Movies)
	}
}

func (h *WebHandler) Ready(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := utils.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := h.roomService.GetRoom(roomName)
	if ok {
		h.roomService.ToggleUserReady(room.Name, user.Username)
	}
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

	user := utils.GetUserFromContext(r)
	if user == nil {
		utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	req.Username = user.Username
	room, ok := h.roomService.GetRoom(req.Room)
	if ok {
		h.roomService.AddMessage(room.Name, req)
	}
}

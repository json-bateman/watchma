package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"watchma/pkg/types"
	"watchma/pkg/utils"
	"watchma/view/rooms"
	"watchma/view/steps"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) SingleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := utils.GetUserFromContext(r)

	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	myRoom, ok := h.services.RoomService.GetRoom(roomName)

	if !ok {
		response := NewPageResponse(rooms.NoRoom(roomName), roomName)
		h.RenderPage(response, w, r)
		return
	}

	if myRoom.Game.MaxPlayers <= len(myRoom.Players) {
		response := NewPageResponse(rooms.RoomFull(), roomName)
		h.RenderPage(response, w, r)
		return
	}

	h.services.RoomService.AddPlayerToRoom(myRoom.Name, user.Username)

	response := NewPageResponse(steps.Lobby(myRoom, user.Username), myRoom.Name)
	h.RenderPage(response, w, r)
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
	myRoom, ok := h.services.RoomService.GetRoom(roomName)
	if ok {
		userBox := steps.UserBox(myRoom, user.Username)
		if err := sse.PatchElementTempl(userBox); err != nil {
			h.logger.Error("Error patching initial user list")
		}
	}

	player, ok := myRoom.GetPlayer(user.Username)
	if !ok {
		h.logger.Error("User not in room", "Username", user.Username, "Room", myRoom.Name)
		return
	}

	// Send existing message history to new client
	if ok {
		if len(myRoom.RoomMessages) > 0 {
			chat := steps.ChatBox(myRoom.RoomMessages)
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
		case utils.ROOM_UPDATE_EVENT:
			userBox := steps.UserBox(myRoom, user.Username)
			if err := sse.PatchElementTempl(userBox); err != nil {
				h.logger.Error("Error patching user list", "error", err)
				return
			}
		case utils.MESSAGE_SENT_EVENT:
			chat := steps.ChatBox(myRoom.RoomMessages)
			if err := sse.PatchElementTempl(chat); err != nil {
				h.logger.Error("Error patching chat message", "error", err)
				return
			}
		case utils.ROOM_START_EVENT:
			draft := steps.Draft(player, movies, h.settings.JellyfinBaseURL, myRoom)
			if err := sse.PatchElementTempl(draft); err != nil {
				h.logger.Error("Error patching movies", "error", err)
				return
			}
		case utils.ROOM_VOTING_EVENT:
			movies := steps.VotingGrid(myRoom.Game.Movies, h.settings.JellyfinBaseURL, myRoom)
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
			finalScreen := steps.ResultsScreen(movieVotes, h.settings.JellyfinBaseURL)
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

	room, ok := h.services.RoomService.GetRoom(roomName)
	if !ok {
		// Room doesn't exist
		utils.WriteJSONError(w, http.StatusNotFound, "Room not found")
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

func (h *WebHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	room, ok := h.services.RoomService.GetRoom(roomName)
	if ok {
		room.Game.Step = types.Draft
		movies, err := h.services.MovieService.GetMovies()
		if err != nil {
			h.logger.Error("Call to MovieService.GetMovies failed", "Error", err)
			return
		}

		if len(movies) == 0 {
			h.logger.Info(fmt.Sprintf("Room %s: No Movies Found", room.Name))
		}

		room.Game.Movies = movies

		h.services.RoomService.StartGame(roomName, room.Game.Movies)
	}
}

func (h *WebHandler) Ready(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := utils.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := h.services.RoomService.GetRoom(roomName)
	if ok {
		h.services.RoomService.TogglePlayerReady(room.Name, user.Username)
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
	room, ok := h.services.RoomService.GetRoom(req.Room)
	if ok {
		h.services.RoomService.AddMessage(room.Name, req)
	}
}

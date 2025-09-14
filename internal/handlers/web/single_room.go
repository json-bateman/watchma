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
	if !ok {
		component := rooms.NoRoom(roomName)
		templ.Handler(component).ServeHTTP(w, r)
		return
	}

	if myRoom.Game.MaxPlayers <= len(myRoom.Users) {
		component := rooms.RoomFull(roomName)
		templ.Handler(component).ServeHTTP(w, r)
		return
	}

	component := rooms.SingleRoom(myRoom, username)
	templ.Handler(component).ServeHTTP(w, r)
}

// Function that does the heavy lifting by keeping the SSE channel open and sending
// Events to the client in real-time
func (h *WebHandler) SingleRoomSSE(w http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "roomName")
	username := utils.GetUsernameFromCookie(r)
	sse := datastar.NewSSE(w, r)
	client := make(chan string, 100)

	h.AddClient(room, client)

	// Send existing user list to new client
	myRoom, ok := h.roomService.GetRoom(room)
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

	// Cleanup when connection closes
	defer func() {
		h.RemoveClient(room, client)
		close(client)
	}()

	for {
		select {
		case message := <-client:
			switch message {
			case utils.ROOM_UPDATE_EVENT:
				userBox := rooms.UserBox(myRoom, username)
				if err := sse.PatchElementTempl(userBox); err != nil {
					fmt.Println("Error patching user list")
					return
				}
			case utils.MESSAGE_SENT_EVENT:
				chat := rooms.ChatBox(myRoom.RoomMessages)
				if err := sse.PatchElementTempl(chat); err != nil {
					fmt.Println("Error patching chat message")
					return
				}
			case utils.ROOM_START_EVENT:
				movies := movies.Movies(myRoom.Game.Movies, h.settings.JellyfinBaseURL, myRoom)
				if err := sse.PatchElementTempl(movies); err != nil {
					fmt.Println("Error patching chat message")
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
					fmt.Println("Error patching chat message")
					return
				}

			default: // discard for now, maybe error?
			}

		case <-r.Context().Done():
			return
		}
	}
}

func (h *WebHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	username := utils.GetUsernameFromCookie(r)

	room, ok := h.roomService.GetRoom(roomName)
	if ok {
		room.AddUser(username)
		h.BroadcastToRoom(roomName, utils.ROOM_UPDATE_EVENT)
		// Broadcast room list update to join page clients
		h.BroadcastToJoinClients(utils.ROOM_LIST_UPDATE_EVENT)
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"ok": true, "update": "user joined room"})
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
	h.mu.RLock()
	if clients, ok := h.sseClients[roomName]; ok {
		for client := range clients {
			select {
			case client <- message:
			default:
				// Client buffer full, skip
			}
		}
	}
	h.mu.RUnlock()
}

func (h *WebHandler) AddClient(roomName string, client chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.sseClients[roomName] == nil {
		h.sseClients[roomName] = make(map[chan string]bool)
	}
	h.sseClients[roomName][client] = true
}

func (h *WebHandler) RemoveClient(roomName string, client chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.sseClients[roomName], client)
	if len(h.sseClients[roomName]) == 0 {
		delete(h.sseClients, roomName)
	}
}

package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/json-bateman/jellyfin-grabber/internal/utils"
	"github.com/json-bateman/jellyfin-grabber/view/rooms"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) SingleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")

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

	component := rooms.SingleRoom(myRoom)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) SingleRoomSSE(w http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")
	sse := datastar.NewSSE(w, r)
	client := make(chan string, 100)

	h.AddClient(room, client)

	// Send existing user list to new client
	myRoom, ok := h.roomService.GetRoom(room)
	if ok {
		userBox := rooms.UserBox(myRoom)
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
				userBox := rooms.UserBox(myRoom)
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
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"ok": true, "update": "user joined room"})
}

func (h *WebHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	username := utils.GetUsernameFromCookie(r)

	room, ok := h.roomService.GetRoom(roomName)
	if ok {
		room.RemoveUser(username)
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

	username := utils.GetUsernameFromCookie(r)
	if username == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, "Username not found")
		return
	}

	req.Username = username
	room, ok := h.roomService.GetRoom(req.Room)
	if ok {
		// Store message
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

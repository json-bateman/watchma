package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/json-bateman/jellyfin-grabber/internal/utils"
	"github.com/json-bateman/jellyfin-grabber/view/rooms"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) Join(w http.ResponseWriter, r *http.Request) {
	component := rooms.JoinPage(h.roomService.Rooms)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) JoinSSE(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	client := make(chan string, 100)

	h.AddJoinClient(client)

	// Send initial room list to new client
	roomList := rooms.RoomListBody(h.roomService.Rooms)
	if err := sse.PatchElementTempl(roomList); err != nil {
		h.logger.Error("Error patching initial room list")
	}

	// Cleanup when connection closes
	defer func() {
		h.RemoveJoinClient(client)
		close(client)
	}()

	for {
		select {
		case message := <-client:
			switch message {
			case utils.ROOM_LIST_UPDATE_EVENT:
				roomList := rooms.RoomListBody(h.roomService.Rooms)
				if err := sse.PatchElementTempl(roomList); err != nil {
					fmt.Println("Error patching room list")
					return
				}
			default: // discard for now, maybe error?
			}

		case <-r.Context().Done():
			return
		}
	}
}

func (h *WebHandler) Host(w http.ResponseWriter, r *http.Request) {
	component := rooms.HostPage("username")
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) HostForm(w http.ResponseWriter, r *http.Request) {
	username := utils.GetUsernameFromCookie(r)
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
		Host:        username,
		Votes:       make(map[*types.JellyfinItem]int),
	})

	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusSeeOther)
}

func (h *WebHandler) BroadcastToJoinClients(message string) {
	h.mu.RLock()
	if clients, ok := h.sseClients["join"]; ok {
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

func (h *WebHandler) AddJoinClient(client chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.sseClients["join"] == nil {
		h.sseClients["join"] = make(map[chan string]bool)
	}
	h.sseClients["join"][client] = true
}

func (h *WebHandler) RemoveJoinClient(client chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.sseClients["join"], client)
	if len(h.sseClients["join"]) == 0 {
		delete(h.sseClients, "join")
	}
}

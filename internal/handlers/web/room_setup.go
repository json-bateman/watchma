package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/json-bateman/jellyfin-grabber/view/rooms"
)

func (h *WebHandler) Join(w http.ResponseWriter, r *http.Request) {
	component := rooms.JoinPage(h.roomService.Rooms)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) Host(w http.ResponseWriter, r *http.Request) {
	component := rooms.HostPage("username")
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) HostForm(w http.ResponseWriter, r *http.Request) {
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
	})

	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusSeeOther)
}

package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/json-bateman/jellyfin-grabber/internal/rooms"
)

func HostForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	// err := r.ParseMultipartForm(1 << 15) // 32KB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	fmt.Printf("%+v", r.Form)

	// Access fields by name
	// TODO: Something with this data, maybe put it in a room struct
	roomName := r.FormValue("roomName")
	moviesStr := r.FormValue("movies")
	movies, err := strconv.Atoi(moviesStr)
	if err != nil {
		http.Error(w, "Movies must be a number", http.StatusBadRequest)
	}
	if rooms.AllRooms.RoomExists(roomName) {
		http.Error(w, "This room name already exists", http.StatusBadRequest)
		return
	}
	rooms.AllRooms.AddRoom(roomName, &rooms.GameSession{MovieNumber: movies})

	// Redirect to /host/room after POST
	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusSeeOther)
}

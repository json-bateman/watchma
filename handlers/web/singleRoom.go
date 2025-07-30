package web

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/rooms"
	"github.com/json-bateman/jellyfin-grabber/view/game"
)

func SingleRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Query())

	roomName := chi.URLParam(r, "roomName")

	var myRoom *rooms.Room
	for a, b := range rooms.AllRooms.Rooms {
		if roomName == a {
			myRoom = b
		}
	}

	component := game.SingleRoom(myRoom)
	templ.Handler(component).ServeHTTP(w, r)
}

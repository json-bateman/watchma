package web

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/game"
	"github.com/json-bateman/jellyfin-grabber/view/rooms"
)

func SingleRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Query())

	roomName := chi.URLParam(r, "roomName")

	var myRoom *game.Room
	for a, b := range game.AllRooms.Rooms {
		if roomName == a {
			myRoom = b
		}
	}

	component := rooms.SingleRoom(myRoom)
	templ.Handler(component).ServeHTTP(w, r)
}

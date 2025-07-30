package web

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/json-bateman/jellyfin-grabber/internal/rooms"
	"github.com/json-bateman/jellyfin-grabber/view/join"
)

func Join(w http.ResponseWriter, r *http.Request) {
	component := join.JoinPage(rooms.AllRooms.Rooms)
	templ.Handler(component).ServeHTTP(w, r)
}

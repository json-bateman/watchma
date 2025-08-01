package web

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/json-bateman/jellyfin-grabber/internal/game"
	"github.com/json-bateman/jellyfin-grabber/view/join"
)

func Join(w http.ResponseWriter, r *http.Request) {
	component := join.JoinPage(game.AllRooms.Rooms)
	templ.Handler(component).ServeHTTP(w, r)
}

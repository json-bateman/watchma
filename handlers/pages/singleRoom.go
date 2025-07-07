package pages

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/json-bateman/jellyfin-grabber/internal/rooms"
	"github.com/json-bateman/jellyfin-grabber/view/game"
)

func SingleRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Query())

	queryParams := r.URL.Query()

	name := queryParams.Get("name")
	moviesStr := queryParams.Get("movies")

	movies, err := strconv.Atoi(moviesStr)
	if err != nil {

	}

	rooms.AllRooms.AddRoom(name, &rooms.GameSession{})

	component := game.SingleRoom(name)
	templ.Handler(component).ServeHTTP(w, r)
}

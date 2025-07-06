package api

import (
	"fmt"
	"net/http"

	"github.com/json-bateman/jellyfin-grabber/view/host"
	"github.com/starfederation/datastar/sdk/go"
)

func HostForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1 << 15) // 32KB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	fmt.Printf("%+v", r.Form)

	// Access fields by name
	// TODO: Something with this data, maybe put it in a room struct
	roomName := r.FormValue("roomName")
	movies := r.FormValue("movies")

	// TODO: Maybe delete this and just redirect user to the room
	sse := datastar.NewSSE(w, r)
	sse.MergeFragmentTempl(host.SubmitButton())

	fmt.Printf("Room Name: %s\nmovies: %s\n", roomName, movies)
	fmt.Fprintf(w, "Room Name: %s\nmovies: %s", roomName, movies)
}

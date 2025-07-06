package api

import (
	"fmt"
	"net/http"
)

func HostForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	fmt.Printf("%+v", r.Form)

	// Access fields by name
	roomName := r.FormValue("roomName")
	movies := r.FormValue("movies")

	fmt.Printf("Room Name: %s\nmovies: %s\n", roomName, movies)
	fmt.Fprintf(w, "Room Name: %s\nmovies: %s", roomName, movies)
}

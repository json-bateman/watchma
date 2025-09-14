package web

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/json-bateman/jellyfin-grabber/internal/utils"
	"github.com/json-bateman/jellyfin-grabber/view/movies"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) Shuffle(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	intVal, err := strconv.Atoi(number)
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "param must be a number", http.StatusBadRequest)
		return
	}

	items, err := h.movieService.FetchJellyfinMovies()
	if err != nil {
		slog.Error("Error fetching jellyfin movies!\n" + err.Error())
		http.Error(w, "Unable to load movies", http.StatusInternalServerError)
		return
	}

	if items == nil || len(items.Items) == 0 {
		log.Printf("no movies found")
	}

	rand.Shuffle(len(items.Items), func(i, j int) {
		items.Items[i], items.Items[j] = items.Items[j], items.Items[i]
	})

	var randMovies []types.JellyfinItem
	if len(items.Items) >= intVal {
		randMovies = items.Items[:intVal]
	} else {
		randMovies = items.Items
	}

	component := movies.Shuffle(randMovies, h.settings.JellyfinBaseURL)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) SubmitMovies(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	username := utils.GetUsernameFromCookie(r)
	var moviesReq types.MovieRequest
	fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&moviesReq); err != nil {
		utils.SendSSEError(w, r, "Invalid Request Body")
		return
	}

	if len(moviesReq.Movies) == 0 {
		utils.SendSSEError(w, r, "Must include at least 1 movie id.")
		return
	}
	room, ok := h.roomService.GetRoom(roomName)
	user, ok2 := room.GetUser(username)
	if ok && ok2 {
		for _, movieID := range moviesReq.Movies {
			// Find the JellyfinItem that matches this ID
			for i := range room.Game.Movies {
				if room.Game.Movies[i].Id == movieID {
					room.Game.Votes[&room.Game.Movies[i]]++
					break
				}
			}
			user.SelectedMovies = append(user.SelectedMovies, movieID)
		}
	}
	user.HasSelectedMovies = true

	sse := datastar.NewSSE(w, r)

	playersComplete := 0
	for _, user := range room.Users {
		if user.HasSelectedMovies {
			playersComplete++
		}
	}
	// If all players have selected movies, push them to final screen
	if playersComplete == len(room.Users) {
		h.BroadcastToRoom(roomName, utils.ROOM_FINISH_EVENT)
	} else {
		// If not, render successfully submitted movies button
		buttonAndMovies := movies.SubmitButton(room.Game.Movies, h.settings.JellyfinBaseURL, user.SelectedMovies)
		sse.PatchElementTempl(buttonAndMovies)
	}
}

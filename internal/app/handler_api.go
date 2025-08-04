package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/json-bateman/jellyfin-grabber/internal"
	"github.com/json-bateman/jellyfin-grabber/internal/game"
	"github.com/json-bateman/jellyfin-grabber/view/movies"
	"github.com/starfederation/datastar-go/datastar"
)

func (a *App) HostForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	// err := r.ParseMultipartForm(1 << 15) // 32KB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Access fields by name
	// TODO: Something with this data, maybe put it in a room struct
	roomName := r.FormValue("roomName")
	moviesStr := r.FormValue("movies")
	movies, err := strconv.Atoi(moviesStr)
	if err != nil {
		http.Error(w, "Movies must be a number", http.StatusBadRequest)
	}
	if game.AllRooms.RoomExists(roomName) {
		http.Error(w, "This room name already exists", http.StatusBadRequest)
		return
	}
	game.AllRooms.AddRoom(roomName, &game.GameSession{MovieNumber: movies})

	// Redirect to /host/room after POST
	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusSeeOther)
}

type movieReq struct {
	MoviesReq []string `json:"movies"`
}

func (a *App) PostMovies(w http.ResponseWriter, r *http.Request) {
	var moviesReq movieReq
	fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&moviesReq); err != nil {
		internal.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}
	if len(moviesReq.MoviesReq) == 0 {
		internal.WriteJSONError(w, http.StatusBadRequest, "Must include at least 1 movie id.")
		return
	}
	// TODO: Something with the movies

	// TODO: Maybe delete this and just redirect user to the room
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(movies.SubmitButton())

	fmt.Println(moviesReq.MoviesReq)
}

type Username struct {
	Username string `json:"username"`
	Roomname string `json:"roomname"`
}

var userSessions = make(map[string]string)

func (a *App) SetUsername(w http.ResponseWriter, r *http.Request) {
	var u Username
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		SendSSEError(w, r, "Bad request", http.StatusBadRequest)
		return
	}

	if u.Username == "" {
		SendSSEError(w, r, "Please Enter a Username", http.StatusBadRequest)
		return
	}

	userSessions[u.Username] = u.Roomname

	fmt.Printf("User %s joined room %s\n", u.Username, u.Roomname)

	ClearSSEError(w, r)
	sse := datastar.NewSSE(w, r)
	if err := sse.PatchElementf(
		`<div id="inputs"> <p>Welcome, %s! You've joined room %s.</p> </div>`, u.Username, u.Roomname,
	); err != nil {
		fmt.Println("there was an error afoot", err)
		return
	}
}

func SendSSEError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElementf(`<div id="error" class="error-message text-red-500">%s</div>`, message)
}

func ClearSSEError(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElements(`<div id="error" class="hidden"></div>`)
}

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/json-bateman/jellyfin-grabber/internal/services"
	"github.com/json-bateman/jellyfin-grabber/internal/utils"
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
	username := r.FormValue("username")
	maxPlayersStr := r.FormValue("maxplayers")

	movies, err := strconv.Atoi(moviesStr)
	maxPlayers, err := strconv.Atoi(maxPlayersStr)
	if err != nil {
		http.Error(w, "Movies must be a number", http.StatusBadRequest)
		return
	}
	if services.AllRooms.RoomExists(roomName) {
		http.Error(w, "This room name already exists", http.StatusBadRequest)
		return
	}
	services.AllRooms.AddRoom(roomName, &services.GameSession{
		MovieNumber: movies,
		MaxPlayers:  maxPlayers,
	})
	room, _ := services.AllRooms.GetRoom(roomName)
	room.AddUser(username)

	// Set username cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jelly_user",
		Value:    username,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		HttpOnly: false,
		Secure:   false,
	})

	// Redirect to room (no username in URL needed)
	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomName), http.StatusSeeOther)
}

type movieReq struct {
	MoviesReq []string `json:"movies"`
}

func (a *App) PostMovies(w http.ResponseWriter, r *http.Request) {
	var moviesReq movieReq
	fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&moviesReq); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}
	if len(moviesReq.MoviesReq) == 0 {
		utils.WriteJSONError(w, http.StatusBadRequest, "Must include at least 1 movie id.")
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

func SendSSEError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElementf(`<div id="error" class="error-message text-red-500">%s</div>`, message)
}

func ClearSSEError(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElements(`<div id="error" class="hidden"></div>`)
}

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

	http.SetCookie(w, &http.Cookie{
		Name:     "jelly_user",
		Value:    u.Username,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		HttpOnly: false,             // Allow JS to read if needed
		Secure:   false,             // Set to true in production with HTTPS
	})

	ClearSSEError(w, r)

	http.Redirect(w, r, "/join", http.StatusSeeOther)
}

// PublishToNATS publishes a JSON payload {"subject": string, "message": string} to the configured NATS server.
type natsPublishRequest struct {
	Subject  string `json:"subject"`
	Message  string `json:"message"`
	Username string `json:"username"`
}

type ChatMessage struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func (a *App) PublishToNATS(w http.ResponseWriter, r *http.Request) {
	if a.Nats == nil {
		utils.WriteJSONError(w, http.StatusServiceUnavailable, "NATS connection not initialized")
		return
	}

	var req natsPublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}
	if req.Subject == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, "Missing subject")
		return
	}

	username := utils.GetUsernameFromCookie(r)
	if username == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, "Username not found in cookie")
		return
	}

	chatMsg := ChatMessage{
		Username: username,
		Message:  req.Message,
	}

	msgBytes, err := json.Marshal(chatMsg)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to encode message")
		return
	}

	if err := a.Nats.Publish(req.Subject, msgBytes); err != nil {
		utils.WriteJSONError(w, http.StatusBadGateway, fmt.Sprintf("Publish failed: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"subject": req.Subject,
	})
}

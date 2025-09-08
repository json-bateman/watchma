package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/json-bateman/jellyfin-grabber/internal/utils"
	"github.com/nats-io/nats.go"
)

// APIHandlers holds dependencies needed by API handlers
type APIHandlers struct {
	Nats         *nats.Conn
	gameClients  map[string]map[chan string]bool
	roomMessages map[string][]string
	mu           *sync.RWMutex
}

func NewAPIHandlers(nats *nats.Conn, gameClients map[string]map[chan string]bool, roomMessages map[string][]string) *APIHandlers {
	return &APIHandlers{
		Nats:         nats,
		gameClients:  gameClients,
		roomMessages: roomMessages,
	}
}

// Sets up all API Routes through Chi Router.
// API Routes should return non-web elements, or perform server actions (I.E. JSON, Send messages to NATS)
func (h *APIHandlers) SetupRoutes(r chi.Router) {

	// Protected API routes
	r.Group(func(r chi.Router) {
		r.Use(RequireUsername)

		r.Post("/nats/publish", h.PublishToNATS)

		// r.Post("/rooms/{roomName}/join", h.JoinRoom)
		// r.Post("/rooms/{roomName}/leave", h.LeaveRoom)
	})
}

func (h *APIHandlers) PublishToNATS(w http.ResponseWriter, r *http.Request) {
	var req types.NatsPublishRequest

	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, "Failed to read request body", http.StatusBadRequest)
	// 	return
	// }
	// defer r.Body.Close()
	//
	// bodyString := string(body)
	// fmt.Println(bodyString)

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
	req.Username = username

	msgBytes, err := json.Marshal(req)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to encode message")
		return
	}

	if err := h.Nats.Publish(req.Subject, msgBytes); err != nil {
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

// func (h *APIHandlers) JoinRoom(w http.ResponseWriter, r *http.Request) {
// 	roomName := chi.URLParam(r, "roomName")
// 	username := utils.GetUsernameFromCookie(r)
//
// 	if username == "" {
// 		http.Error(w, "No username found", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Create structured message
// 	msg := types.RoomMessage{
// 		Subject:  utils.JOIN_MSG,
// 		Username: username,
// 	}
//
// 	msgBytes, err := json.Marshal(msg)
// 	if err != nil {
// 		http.Error(w, "Failed to create message", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Broadcast to all clients in this room
// 	h.mu.RLock()
// 	if clients, ok := h.gameClients[roomName]; ok {
// 		for client := range clients {
// 			select {
// 			case client <- string(msgBytes):
// 			default:
// 				// Client buffer full, skip
// 			}
// 		}
// 	}
// 	h.mu.RUnlock()
//
// 	w.WriteHeader(http.StatusOK)
// }
//
// func (h *APIHandlers) LeaveRoom(w http.ResponseWriter, r *http.Request) {
// 	roomName := chi.URLParam(r, "roomName")
// 	username := utils.GetUsernameFromCookie(r)
//
// 	if username == "" {
// 		http.Error(w, "No username found", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Create structured message
// 	msg := types.RoomMessage{
// 		Subject:  utils.LEAVE_MSG,
// 		Username: username,
// 	}
//
// 	msgBytes, err := json.Marshal(msg)
// 	if err != nil {
// 		http.Error(w, "Failed to create message", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Broadcast to all clients in this room
// 	h.mu.RLock()
// 	if clients, ok := h.gameClients[roomName]; ok {
// 		for client := range clients {
// 			select {
// 			case client <- string(msgBytes):
// 			default:
// 				// Client buffer full, skip
// 			}
// 		}
// 	}
// 	h.mu.RUnlock()
//
// 	w.WriteHeader(http.StatusOK)
// }

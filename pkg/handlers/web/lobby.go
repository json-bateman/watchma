package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"watchma/pkg/types"

	"github.com/go-chi/chi/v5"
)

func (h *WebHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	room, ok := h.services.RoomService.GetRoom(roomName)
	if ok {
		room.Game.Step = types.Draft
		movies, err := h.services.MovieService.GetMovies()
		if err != nil {
			h.logger.Error("Call to MovieService.GetMovies failed", "Error", err)
			return
		}

		if len(movies) == 0 {
			h.logger.Info(fmt.Sprintf("Room %s: No Movies Found", room.Name))
		}

		room.Game.Movies = movies

		h.services.RoomService.StartGame(roomName, room.Game.Movies)
	}
}

func (h *WebHandler) Ready(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := h.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := h.services.RoomService.GetRoom(roomName)
	if ok {
		h.services.RoomService.TogglePlayerReady(room.Name, user.Username)
	}
}

func (h *WebHandler) PublishChatMessage(w http.ResponseWriter, r *http.Request) {
	var req types.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}
	if len(strings.Trim(req.Message, " ")) == 0 {
		return
	}

	user := h.GetUserFromContext(r)
	if user == nil {
		h.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	req.Username = user.Username
	room, ok := h.services.RoomService.GetRoom(req.Room)
	if ok {
		h.services.RoomService.AddMessage(room.Name, req)
	}
}

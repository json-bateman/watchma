package services

import (
	"fmt"
	"log/slog"

	"github.com/json-bateman/jellyfin-grabber/internal/game"
)

type RoomService struct {
	roomManager *game.RoomManager
	logger      *slog.Logger
}

func NewRoomService(roomManager *game.RoomManager, logger *slog.Logger) *RoomService {
	return &RoomService{
		roomManager: roomManager,
		logger:      logger,
	}
}

func (rs *RoomService) CreateRoom(name string, hostUserID string, movieCount int) (*game.Room, error) {
	if rs.roomManager.RoomExists(name) {
		return nil, fmt.Errorf("room %s already exists", name)
	}

	gameSession := &game.GameSession{
		Movies:      []string{},
		MovieNumber: movieCount,
		Votes:       make(map[string]int),
		Step:        game.Lobby,
	}

	rs.roomManager.AddRoom(name, gameSession)
	rs.logger.Info("Room created",
		"room_name", name,
		"host_user_id", hostUserID,
		"movie_count", movieCount,
	)

	room, _ := rs.roomManager.GetRoom(name)
	return room, nil
}

func (rs *RoomService) GetOrCreateRoom(name string) (*game.Room, error) {
	if room, exists := rs.roomManager.GetRoom(name); exists {
		return room, nil
	}

	return rs.CreateRoom(name, "", 20)
}

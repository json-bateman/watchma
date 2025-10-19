package services

import (
	"log/slog"
	"sort"
	"sync"
	"time"

	"watchma/pkg/types"
	"watchma/pkg/utils"
)

// All mutation and publish events are done by the RoomService
type RoomService struct {
	mu     sync.RWMutex
	Rooms  map[string]*Room
	pub    *EventPublisher
	logger *slog.Logger
}

// Room is a PURE DATA STRUCTURE (no mutation methods)
type Room struct {
	Name         string
	Game         *types.GameSession
	RoomMessages []types.Message
	Players      map[string]*types.Player
	mu           sync.RWMutex
}

func NewRoomService(pub *EventPublisher, l *slog.Logger) *RoomService {
	return &RoomService{
		Rooms:  make(map[string]*Room),
		pub:    pub,
		logger: l,
	}
}

func (rs *RoomService) AddRoom(roomName string, game *types.GameSession) {
	rs.mu.Lock()
	rs.Rooms[roomName] = &Room{
		Name:         roomName,
		Game:         game,
		RoomMessages: make([]types.Message, 0),
		Players:      make(map[string]*types.Player),
	}
	rs.mu.Unlock()

	rs.pub.PublishLobbyEvent(utils.ROOM_LIST_UPDATE_EVENT)
}

func (rs *RoomService) DeleteRoom(roomName string) {
	rs.mu.Lock()
	delete(rs.Rooms, roomName)
	rs.mu.Unlock()

	rs.pub.PublishLobbyEvent(utils.ROOM_LIST_UPDATE_EVENT)
}

func (rs *RoomService) AddPlayerToRoom(roomName, username string) (*types.Player, bool) {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return nil, false
	}

	room.mu.Lock()
	player := &types.Player{
		Username: username,
		JoinedAt: time.Now(),
	}
	room.Players[username] = player
	room.mu.Unlock()

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	rs.pub.PublishLobbyEvent(utils.ROOM_LIST_UPDATE_EVENT)

	return player, true
}

func (rs *RoomService) RemovePlayerFromRoom(roomName, username string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	delete(room.Players, username)
	room.mu.Unlock()

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	rs.pub.PublishLobbyEvent(utils.ROOM_LIST_UPDATE_EVENT)

	return true
}

func (rs *RoomService) TransferHost(roomName, newHost string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	room.Game.Host = newHost
	room.mu.Unlock()

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	return true
}

func (rs *RoomService) TogglePlayerReady(roomName, username string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	player, found := room.Players[username]
	if !found {
		room.mu.Unlock()
		return false
	}
	player.Ready = !player.Ready
	room.mu.Unlock()

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	return true
}

func (rs *RoomService) TogglePlayerFinishedDraft(roomName, username string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	player, found := room.Players[username]
	if !found {
		room.mu.Unlock()
		return false
	}
	player.HasFinishedDraft = !player.HasFinishedDraft
	room.mu.Unlock()

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	return true
}

func (rs *RoomService) AddMessage(roomName string, msg types.Message) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	room.RoomMessages = append(room.RoomMessages, msg)
	room.mu.Unlock()

	rs.pub.PublishRoomEvent(roomName, utils.MESSAGE_SENT_EVENT)
	return true
}

func (rs *RoomService) StartGame(roomName string, movies []types.JellyfinItem) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	room.Game.Step = types.Voting
	room.Game.Movies = movies
	room.mu.Unlock()

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_START_EVENT)
	return true
}

func (rs *RoomService) FinishGame(roomName string) bool {
	_, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_FINISH_EVENT)
	return true
}

func (rs *RoomService) RoomExists(roomName string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	_, exists := rs.Rooms[roomName]
	return exists
}

func (rs *RoomService) GetRoom(roomName string) (*Room, bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	room, ok := rs.Rooms[roomName]
	return room, ok
}

func (r *Room) GetPlayer(username string) (*types.Player, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	player, exists := r.Players[username]
	return player, exists
}

func (r *Room) GetAllPlayers() []*types.Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	players := make([]*types.Player, 0, len(r.Players))
	for _, player := range r.Players {
		players = append(players, player)
	}
	return players
}

func (r *Room) PlayersByJoinTime() []*types.Player {
	r.mu.RLock()
	players := make([]*types.Player, 0, len(r.Players))
	for _, u := range r.Players {
		players = append(players, u)
	}
	r.mu.RUnlock()

	sort.Slice(players, func(i, j int) bool {
		return players[i].JoinedAt.Before(players[j].JoinedAt)
	})
	return players
}

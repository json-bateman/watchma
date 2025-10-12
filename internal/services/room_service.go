package services

import (
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/json-bateman/jellyfin-grabber/internal/types"
	"github.com/json-bateman/jellyfin-grabber/internal/utils"
)

// All mutation and publish events are done by the RoomService
type RoomService struct {
	mu     sync.RWMutex
	Rooms  map[string]*Room
	pub    *EventPublisher
	logger *slog.Logger
}

// Room is now a PURE DATA STRUCTURE (no mutation methods)
type Room struct {
	Name         string
	Game         *types.GameSession
	RoomMessages []types.Message
	Users        map[string]*types.User
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
		Users:        make(map[string]*types.User),
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

func (rs *RoomService) AddUserToRoom(roomName, username string) (*types.User, bool) {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return nil, false
	}

	room.mu.Lock()
	user := &types.User{
		Name:     username,
		JoinedAt: time.Now(),
	}
	room.Users[username] = user
	room.mu.Unlock()

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	rs.pub.PublishLobbyEvent(utils.ROOM_LIST_UPDATE_EVENT)

	return user, true
}

func (rs *RoomService) RemoveUserFromRoom(roomName, username string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	delete(room.Users, username)
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

func (rs *RoomService) ToggleUserReady(roomName, username string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	user, found := room.Users[username]
	if !found {
		room.mu.Unlock()
		return false
	}
	user.Ready = !user.Ready
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

func (r *Room) GetUser(username string) (*types.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.Users[username]
	return user, exists
}

func (r *Room) GetAllUsers() []*types.User {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]*types.User, 0, len(r.Users))
	for _, user := range r.Users {
		users = append(users, user)
	}
	return users
}

func (r *Room) UsersByJoinTime() []*types.User {
	r.mu.RLock()
	users := make([]*types.User, 0, len(r.Users))
	for _, u := range r.Users {
		users = append(users, u)
	}
	r.mu.RUnlock()

	sort.Slice(users, func(i, j int) bool {
		return users[i].JoinedAt.Before(users[j].JoinedAt)
	})
	return users
}

package services

import (
	"sync"
	"time"

	"github.com/json-bateman/jellyfin-grabber/internal/types"
)

type RoomService struct {
	mu    sync.RWMutex
	Rooms map[string]*Room
}

type Room struct {
	Name         string
	Game         *types.GameSession
	RoomMessages []types.Message
	Users        map[string]*types.User // username -> User
	mu           sync.RWMutex
}

func NewRoomService() *RoomService {
	return &RoomService{
		Rooms: make(map[string]*Room),
	}
}

func (rm *RoomService) AddRoom(name string, game *types.GameSession) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.Rooms[name] = &Room{
		Name:         name,
		Game:         game,
		RoomMessages: make([]types.Message, 0),
		Users:        make(map[string]*types.User),
	}
}

func (rm *RoomService) RoomExists(name string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	_, exists := rm.Rooms[name]
	return exists
}

func (rm *RoomService) DeleteRoom(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.Rooms, name)
}

func (rm *RoomService) GetRoom(name string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	room, ok := rm.Rooms[name]
	return room, ok
}

func (r *Room) AddUser(username string) *types.User {
	r.mu.Lock()
	defer r.mu.Unlock()

	user := &types.User{
		Name:     username,
		JoinedAt: time.Now(),
	}
	r.Users[username] = user
	return user
}

func (r *Room) RemoveUser(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Users, username)
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

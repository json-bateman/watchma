package services

import (
	"sort"
	"sync"
	"time"

	"github.com/json-bateman/jellyfin-grabber/internal/types"
)

// RoomService manages all room operations
type RoomService struct {
	mu    sync.RWMutex
	Rooms map[string]*Room
}

// Room manages all operations within rooms, including managing GameSession
type Room struct {
	Name         string
	Game         *types.GameSession
	RoomMessages []types.Message
	Users        map[string]*types.User // username -> User
	mu           sync.RWMutex
}

// NewRoomService creates a RoomService
func NewRoomService() *RoomService {
	return &RoomService{
		Rooms: make(map[string]*Room),
	}
}

// AddRoom adds Room with [roomName] as key to Rooms map with mutex lock
func (rm *RoomService) AddRoom(roomName string, game *types.GameSession) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.Rooms[roomName] = &Room{
		Name:         roomName,
		Game:         game,
		RoomMessages: make([]types.Message, 0),
		Users:        make(map[string]*types.User),
	}
}

// RoomExists checks if [roomName] is in Rooms map with mutex lock
func (rm *RoomService) RoomExists(roomName string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	_, exists := rm.Rooms[roomName]
	return exists
}

// DeleteRoom deletes room from Rooms map with mutex lock
func (rm *RoomService) DeleteRoom(roomName string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.Rooms, roomName)
}

// GetRoom gets room from Rooms map with mutex lock
func (rm *RoomService) GetRoom(roomName string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	room, ok := rm.Rooms[roomName]
	return room, ok
}

// AddUser adds a user to Room.Users map with mutex lock
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

// RemoveUser removes a user to Room.Users map with mutex lock
func (r *Room) RemoveUser(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Users, username)
}

// GetUser gets a user from Room.Users map with mutex lock
func (r *Room) GetUser(username string) (*types.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.Users[username]
	return user, exists
}

// GetAllUsers gets all users from Room.Users map with mutex lock
func (r *Room) GetAllUsers() []*types.User {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]*types.User, 0, len(r.Users))
	for _, user := range r.Users {
		users = append(users, user)
	}
	return users
}

// UsersByJoinTime gets all users by Join Time from Room.Users and sorts them, returning the sorted array.
func (r *Room) UsersByJoinTime() []*types.User {
	r.mu.RLock()
	users := make([]*types.User, 0, len(r.Users))
	for _, u := range r.Users {
		users = append(users, u) // copy values to avoid races on fields
	}
	r.mu.RUnlock()

	sort.Slice(users, func(i, j int) bool {
		return users[i].JoinedAt.Before(users[j].JoinedAt)
	})
	return users
}

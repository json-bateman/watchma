package services

import (
	"sync"
	"time"
)

type GameStep int

const (
	Lobby = iota
	Movies
)

type GameSession struct {
	Movies      []string
	MovieNumber int
	MaxPlayers  int
	Votes       map[string]int // Movie -> vote count
	Step        GameStep
}

type User struct {
	Name     string
	JoinedAt time.Time
}

// ----- Structs with Methods -----
type RoomManager struct {
	mu    sync.RWMutex
	Rooms map[string]*Room
}

type Room struct {
	Name  string
	Game  *GameSession
	Users map[string]*User // username -> User
	mu    sync.RWMutex
}

var AllRooms = &RoomManager{
	Rooms: make(map[string]*Room),
}

func (rm *RoomManager) AddRoom(name string, game *GameSession) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.Rooms[name] = &Room{
		Name:  name,
		Game:  game,
		Users: make(map[string]*User),
	}
}

func (rm *RoomManager) RoomExists(name string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	_, exists := rm.Rooms[name]
	return exists
}

func (rm *RoomManager) DeleteRoom(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.Rooms, name)
}

func (rm *RoomManager) GetRoom(name string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	room, ok := rm.Rooms[name]
	return room, ok
}

func (r *Room) AddUser(username string) *User {
	r.mu.Lock()
	defer r.mu.Unlock()

	user := &User{
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

func (r *Room) GetUser(username string) (*User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.Users[username]
	return user, exists
}

func (r *Room) GetAllUsers() []*User {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]*User, 0, len(r.Users))
	for _, user := range r.Users {
		users = append(users, user)
	}
	return users
}

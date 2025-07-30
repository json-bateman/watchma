package rooms

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type GameStep int

type User struct {
	ID       string
	Name     string
	Conn     *websocket.Conn
	JoinedAt time.Time
}

type Message struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type Room struct {
	Name  string
	Game  *GameSession
	Users map[string]*User
	mu    sync.RWMutex
}

type GameSession struct {
	Movies      []string       // The n movie IDs to choose from
	MovieNumber int            // amount of movies to randomize and show to user
	Votes       map[string]int // Movie -> vote count
	Step        GameStep       // Game State Step

	// Users       map[string]*User // UserID -> User info, optional?
}

type RoomManager struct {
	mu    sync.RWMutex
	Rooms map[string]*Room
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

package rooms

import (
	"sync"
)

type GameStep int

const (
	Lobby GameStep = iota
	Voting
	Results
	Finished
)

func (s GameStep) String() string {
	switch s {
	case Lobby:
		return "lobby"
	case Voting:
		return "voting"
	case Results:
		return "results"
	case Finished:
		return "finished"
	default:
		return "unknown"
	}
}

// type User struct {
// 	ID   string
// 	Name string
// }

type Room struct {
	Name string
	Game *GameSession
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

// Mutexes for concurrency, might be a pre-optimization tbh
func (rm *RoomManager) AddRoom(name string, game *GameSession) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.Rooms[name] = &Room{Name: name, Game: game}
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

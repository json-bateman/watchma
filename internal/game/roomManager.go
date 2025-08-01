package game

import "sync"

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

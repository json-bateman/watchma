package game

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
	Votes       map[string]int // Movie -> vote count
	Step        GameStep
}

type User struct {
	Name     string
	JoinedAt time.Time
}

type Room struct {
	Name  string
	Game  *GameSession
	Users map[string]*User // username -> User
	mu    sync.RWMutex
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

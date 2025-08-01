package game

import (
	"sync"
	"time"
)

type GameStep int

const (
	Lobby = iota
)

type GameSession struct {
	Movies      []string
	MovieNumber int
	Votes       map[string]int // Movie -> vote count
	Step        GameStep

	// Users       map[string]*User
}

type User struct {
	ID       string
	Name     string
	JoinedAt time.Time
}

type Message struct {
	Type      string         `json:"type"`
	Data      map[string]any `json:"data"`
	Timestamp time.Time      `json:"timestamp"`
}

type Room struct {
	Name  string
	Game  *GameSession
	Users map[string]*User
	mu    sync.RWMutex
}

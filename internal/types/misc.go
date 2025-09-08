package types

import "time"

// Dumping unorganized types in this file until there's enough to refactor

type GameStep int

const (
	Lobby = iota
	Movies
)

type GameSession struct {
	Movies      []string
	MovieNumber int
	MaxPlayers  int
	Votes       map[string]int // MovieID -> vote count
	Step        GameStep
}

// MovieRequest represents a request containing movie IDs
type MovieRequest struct {
	MoviesReq []string `json:"moviesReq"`
}

type User struct {
	Name     string
	JoinedAt time.Time
}

// Message represents a chat message
type Message struct {
	Subject  string `json:"subject"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Room     string `json:"room"`
}

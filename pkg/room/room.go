package room

import (
	"sync"
	"time"

	"watchma/pkg/movie"
)

// Room represents a single room for players, cleans up when all players leave
type Room struct {
	Name         string
	Game         *Session
	RoomMessages []Message
	Players      map[string]*Player
	mu           sync.RWMutex
}

type Player struct {
	Username          string
	JoinedAt          time.Time
	Ready             bool
	DraftMovies       []movie.Movie
	VotingMovies      []movie.Movie
	HasFinishedDraft  bool
	HasFinishedVoting bool
}

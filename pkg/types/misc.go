package types

import (
	"time"
)

// Dumping unorganized types in this file until there's enough to refactor

type GameStep int

const (
	Lobby = iota
	Draft
	Voting
	Results
)

type GameSession struct {
	Host          string
	Movies        []JellyfinItem
	MovieNumber   int
	MaxPlayers    int
	MaxDraftCount int
	Votes         map[*JellyfinItem]int // MovieID -> vote count
	Step          GameStep
}

// MovieRequest represents a request containing movie IDs
type MovieRequest struct {
	Movies []string `json:"movies"`
}

type Player struct {
	Username          string
	JoinedAt          time.Time
	Ready             bool
	DraftedMovies     []string // MovieId
	SelectedMovies    []string // MovieId
	HasFinishedDraft  bool
	HasSelectedMovies bool
}

// Message represents a chat message
type Message struct {
	Subject  string `json:"subject"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Room     string `json:"room"`
}

// MovieVote represents a struct for holding final votes
type MovieVote struct {
	Movie *JellyfinItem
	Votes int
}

type DraftState struct {
	SelectedMovies []JellyfinItem
	IsReady        bool
	MaxVotes       int
}

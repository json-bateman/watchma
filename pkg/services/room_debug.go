package services

import (
	"watchma/pkg/types"
)

// RoomDebugInfo contains formatted debug information
type RoomDebugInfo struct {
	RoomName      string
	Step          string
	Host          string
	PlayerCount   int
	Players       []PlayerDebugInfo
	DraftMovies   []types.Movie
	VotingMovies  []types.Movie
	MaxPlayers    int
	MaxDraftCount int
}

type PlayerDebugInfo struct {
	Username          string
	Ready             bool
	DraftMovies       int
	VotingMovies      int
	HasFinishedDraft  bool
	HasSelectedMovies bool
}

// GetDebugSnapshot returns a snapshot of all room states
func (rs *RoomService) GetDebugSnapshot() []RoomDebugInfo {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	snapshot := make([]RoomDebugInfo, 0, len(rs.Rooms))

	for _, room := range rs.Rooms {
		room.mu.RLock()

		players := make([]PlayerDebugInfo, 0, len(room.Players))
		for _, p := range room.Players {
			players = append(players, PlayerDebugInfo{
				Username:          p.Username,
				Ready:             p.Ready,
				DraftMovies:       len(p.DraftMovies),
				VotingMovies:      len(p.VotingMovies),
				HasFinishedDraft:  p.HasFinishedDraft,
				HasSelectedMovies: p.HasFinishedVoting,
			})
		}

		snapshot = append(snapshot, RoomDebugInfo{
			RoomName:      room.Name,
			Step:          getStepName(room.Game.Step),
			Host:          room.Game.Host,
			PlayerCount:   len(room.Players),
			MaxPlayers:    room.Game.MaxPlayers,
			MaxDraftCount: room.Game.MaxDraftCount,
			Players:       players,
			VotingMovies:  room.Game.VotingMovies,
			DraftMovies:   room.Game.VotingMovies,
		})

		room.mu.RUnlock()
	}

	return snapshot
}

func getStepName(step types.GameStep) string {
	switch step {
	case types.Lobby:
		return "Lobby"
	case types.Draft:
		return "Draft"
	case types.Voting:
		return "Voting"
	case types.Results:
		return "Results"
	default:
		return "Unknown"
	}
}

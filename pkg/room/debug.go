package room

import (
	"watchma/pkg/movie"
)

// DebugInfo contains formatted debug information
type DebugInfo struct {
	RoomName      string
	Step          string
	Host          string
	PlayerCount   int
	Players       []PlayerDebug
	VotingMovies  []movie.Movie
	MaxPlayers    int
	MaxDraftCount int
}

type PlayerDebug struct {
	Username          string
	Ready             bool
	DraftMovies       int
	VotingMovies      int
	HasFinishedDraft  bool
	HasSelectedMovies bool
	AvailableMovies   []movie.Movie
}

// GetDebugSnapshot returns a snapshot of all room states
func (rs *Service) GetDebugSnapshot() []DebugInfo {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	snapshot := make([]DebugInfo, 0, len(rs.Rooms))

	for _, room := range rs.Rooms {
		room.mu.RLock()

		players := make([]PlayerDebug, 0, len(room.Players))
		for _, p := range room.Players {
			players = append(players, PlayerDebug{
				Username:          p.Username,
				Ready:             p.Ready,
				DraftMovies:       len(p.DraftMovies),
				VotingMovies:      len(p.VotingMovies),
				HasFinishedDraft:  p.HasFinishedDraft,
				HasSelectedMovies: p.HasFinishedVoting,
				AvailableMovies:   p.AvailableMovies,
			})
		}

		snapshot = append(snapshot, DebugInfo{
			RoomName:      room.Name,
			Step:          getStepName(room.Game.Step),
			Host:          room.Game.Host,
			PlayerCount:   len(room.Players),
			MaxPlayers:    room.Game.MaxPlayers,
			MaxDraftCount: room.Game.MaxDraftCount,
			Players:       players,
			VotingMovies:  room.Game.VotingMovies,
		})

		room.mu.RUnlock()
	}

	return snapshot
}

func getStepName(step Step) string {
	switch step {
	case Lobby:
		return "Lobby"
	case Draft:
		return "Draft"
	case Voting:
		return "Voting"
	case Results:
		return "Results"
	default:
		return "Unknown"
	}
}

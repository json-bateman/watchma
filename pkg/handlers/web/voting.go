package web

import (
	"net/http"
	"sort"
	"watchma/pkg/services"
	"watchma/pkg/types"
	"watchma/view/steps"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *WebHandler) VotingSubmit(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	room, ok := h.services.RoomService.GetRoom(roomName)

	currentUser := h.GetUserFromContext(r)
	if currentUser == nil {
		h.logger.Error("No User found from session cookie")
		return
	}

	player, ok := room.GetPlayer(currentUser.Username)
	if !ok {
		h.logger.Error("Player not in room")
		return
	}
	if len(player.VotingMovies) == 0 {
		h.SendSSEError(w, r, "Must include at least 1 movie id.")
		return
	}

	player.HasFinishedVoting = true

	isVotingFinished := room.IsVotingFinished()

	// if voting is finished, add all players choices to the voting array
	if isVotingFinished {
		room.Game.Step = types.Results
		h.services.RoomService.SubmitFinalVotes(room)
		h.services.RoomService.AnnounceWinner(roomName)
	} else {
		h.renderVotingPage(w, r)
	}
}

func (h *WebHandler) ToggleVotingMovie(w http.ResponseWriter, r *http.Request) {
	movieId := chi.URLParam(r, "id")
	user := h.GetUserFromContext(r)
	roomName := chi.URLParam(r, "roomName")
	room, _ := h.services.RoomService.GetRoom(roomName)

	movie := room.Game.AllMoviesMap[movieId]

	if !h.services.RoomService.ToggleVotingMovie(roomName, user.Username, *movie) {
		h.logger.Error("Failed to toggle steps.Voting movie", "Room", roomName, "Username", user.Username, "MovieId", movieId)
	}

	h.renderVotingPage(w, r)
}

func (h *WebHandler) renderVotingPage(w http.ResponseWriter, r *http.Request) {
	// Voting page needs all movies, and the room
	roomName := chi.URLParam(r, "roomName")
	room, ok := h.services.RoomService.GetRoom(roomName)

	if !ok {
		h.logger.Error("Could not obtain room", "room", roomName)
		return
	}

	currentUser := h.GetUserFromContext(r)
	if currentUser == nil {
		h.logger.Error("No User found from session cookie")
		return
	}

	player, ok := room.GetPlayer(currentUser.Username)
	if !ok {
		h.logger.Error("Player not in room")
		return
	}

	draft := steps.Voting(room.Game.VotingMovies, player, room)
	if err := datastar.NewSSE(w, r).PatchElementTempl(draft); err != nil {
		h.logger.Error("Error Rendering Draft Page")
	}
}

// SortMoviesByVotes converts a vote map to a sorted slice (descending by votes)
func SortMoviesByVotes(votes map[*types.Movie]int) []types.MovieVote {
	movieVotes := make([]types.MovieVote, 0, len(votes))

	for movie, voteCount := range votes {
		movieVotes = append(movieVotes, types.MovieVote{
			Movie: movie,
			Votes: voteCount,
		})
	}

	// Sort by votes (descending - highest votes first)
	sort.Slice(movieVotes, func(i, j int) bool {
		return movieVotes[i].Votes > movieVotes[j].Votes
	})

	return movieVotes
}

// GetWinnerMovies returns only movies with the highest vote count
// Will return a randomly selected movie if display ties is false
// Since SortMoviesByVotes is ranging over a map, the selection is random
func GetWinnerMovies(moviesSortedByVote []types.MovieVote, room *services.Room) []types.MovieVote {
	if len(moviesSortedByVote) == 0 {
		return []types.MovieVote{}
	}

	if !room.Game.DisplayTies {
		return []types.MovieVote{moviesSortedByVote[0]}
	}

	maxVotes := moviesSortedByVote[0].Votes
	var winners []types.MovieVote

	for _, movie := range moviesSortedByVote {
		if movie.Votes == maxVotes {
			winners = append(winners, movie)
		}
	}

	return winners
}

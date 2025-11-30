package room

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"watchma/db/sqlcgen"
	"watchma/pkg/movie"
)

// Service represents the orchestrator of all rooms and room operations
type Service struct {
	mu      sync.RWMutex
	Rooms   map[string]*Room
	pub     *EventPublisher
	queries *sqlcgen.Queries
	logger  *slog.Logger
}

func NewService(queries *sqlcgen.Queries, pub *EventPublisher, l *slog.Logger) *Service {
	return &Service{
		Rooms:   make(map[string]*Room),
		pub:     pub,
		queries: queries,
		logger:  l,
	}
}

func (rs *Service) AddRoom(roomName string, game *Session) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Rooms[roomName] = &Room{
		Name:         roomName,
		Game:         game,
		RoomMessages: make([]Message, 0),
		Players:      make(map[string]*Player),
	}

	rs.logger.Info("Room added", "name", roomName)

	rs.pub.PublishLobbyEvent(RoomListUpdateEvent)
}

func (rs *Service) DeleteRoom(roomName string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	delete(rs.Rooms, roomName)

	rs.logger.Info("Room deleted", "name", roomName)

	rs.pub.PublishLobbyEvent(RoomListUpdateEvent)
}

func (rs *Service) AddPlayerToRoom(roomName, username string) (*Player, bool) {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return nil, false
	}

	room.mu.Lock()
	defer room.mu.Unlock()
	player := &Player{
		Username: username,
		JoinedAt: time.Now(),
	}
	room.Players[username] = player

	rs.logger.Debug("Player added to room", "roomName", roomName, "playerName", username)

	rs.pub.PublishRoomEvent(roomName, RoomUpdateEvent)
	rs.pub.PublishLobbyEvent(RoomListUpdateEvent)

	return player, true
}

func (rs *Service) RemovePlayerFromRoom(roomName, username string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	defer room.mu.Unlock()
	delete(room.Players, username)

	if len(room.Players) == 0 {
		rs.DeleteRoom(room.Name)
	}

	rs.logger.Debug("Player removed from room", "roomName", roomName, "playerName", username)

	rs.pub.PublishRoomEvent(roomName, RoomUpdateEvent)
	rs.pub.PublishLobbyEvent(RoomListUpdateEvent)

	return true
}

func (rs *Service) TransferHost(roomName, newHost string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	room.Game.Host = newHost
	room.mu.Unlock()

	rs.pub.PublishRoomEvent(roomName, RoomUpdateEvent)
	return true
}

func (rs *Service) TogglePlayerReady(roomName, username string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	defer room.mu.Unlock()
	player, found := room.Players[username]
	if !found {
		return false
	}
	player.Ready = !player.Ready

	rs.pub.PublishRoomEvent(roomName, RoomUpdateEvent)
	return true
}

func (rs *Service) AddMessage(roomName string, msg Message) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	defer room.mu.Unlock()
	room.RoomMessages = append(room.RoomMessages, msg)

	rs.pub.PublishRoomEvent(roomName, MessageSentEvent)
	return true
}

func (rs *Service) StartGame(roomName string, movies []movie.Movie) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	defer room.mu.Unlock()
	room.Game.Step = Draft
	room.Game.SetAllMovies(movies)

	// Give each player their own copy of the movies
	for _, player := range room.Players {
		player.AvailableMovies = movie.CopySlice(movies)
	}

	rs.logger.Info("Game Started", "roomName", roomName)

	rs.pub.PublishRoomEvent(roomName, RoomStartEvent)
	return true
}

func (rs *Service) MoveToVoting(roomName string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	defer room.mu.Unlock()
	room.Game.Step = Voting

	rs.logger.Info("Game Moved to Voting", "roomName", roomName)

	rs.pub.PublishRoomEvent(roomName, RoomVotingEvent)
	return true
}

func (rs *Service) AnnounceWinner(roomName string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.Game.Step = Announce
	rs.logger.Info("Announcing Winner...", "roomName", roomName)

	rs.pub.PublishRoomEvent(roomName, RoomAnnounceEvent)
	return true
}

func (rs *Service) FinishGame(roomName string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.Game.Step = Results
	rs.logger.Info("Game Finished", "roomName", roomName)

	// Save Game result and participants to DB
	go func() {
		if err := rs.SaveGameResult(roomName); err != nil {
			rs.logger.Error("Failed to save game result", "error", err, "room", roomName)
		}
	}()

	rs.pub.PublishRoomEvent(roomName, RoomFinishEvent)
	return true
}

func (rs *Service) RoomExists(roomName string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	_, exists := rs.Rooms[roomName]
	return exists
}

func (rs *Service) GetRoom(roomName string) (*Room, bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	room, ok := rs.Rooms[roomName]
	return room, ok
}

// SubmitDraftVotes() submits all votes for the given room.
// When voting is concluded, we can advance the step
func (rs *Service) SubmitDraftVotes(room *Room) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	for _, p := range room.Players {
		// Add each player's draft movies to voting movies (if not already present)
		for _, movie := range p.DraftMovies {
			if room.Game.VotingMoviesContains(movie) {
				continue
			}

			if movie, exists := room.Game.GetMovie(movie); exists {
				room.Game.VotingMovies = append(room.Game.VotingMovies, *movie)
			}
		}
	}

	rs.logger.Debug("All Draft Votes Submitted to Voting Array", "Room Name", room.Name)
}

func (rs *Service) SubmitFinalVotes(room *Room) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	for _, p := range room.Players {
		// Add up total votes from all players
		for _, m := range p.VotingMovies {
			if moviePtr, ok := room.Game.AllMoviesMap[m.Id]; ok {
				room.Game.Votes[moviePtr]++
			} else {
				rs.logger.Warn("Movie not found in AllMoviesMap", "movieId", m.Id)
			}
		}
	}

	rs.logger.Debug("All Movie Votes submitted to Voting Movies Results Array", "Room Name", room.Name, "votes", room.Game.Votes)
}

// RemoveDraftMovie removes a specific movie from a player's draft selection
// Returns true if the movie was found and removed
func (rs *Service) RemoveDraftMovie(roomName, username string, movieId string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	player, ok := room.Players[username]
	if !ok {
		return false
	}

	for i, m := range player.DraftMovies {
		if m.Id == movieId {
			player.DraftMovies = append(
				player.DraftMovies[:i],
				player.DraftMovies[i+1:]...,
			)
			rs.logger.Debug("Movie removed from draft", "roomName", roomName, "player", username, "movie", m.Name)
			return true
		}
	}

	return false
}

// ToggleDraftMovie adds or removes a movie from a player's draft selection
// If the movie is already in the draft, it will be removed
// If the movie is not in the draft and the player hasn't reached MaxDraftCount, it will be added
// Returns true if toggle occurred
func (rs *Service) ToggleDraftMovie(roomName, username string, movie movie.Movie) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	player, ok := room.Players[username]
	if !ok {
		return false
	}
	var wasToggled bool
	var action string

	// Try to remove the movie if it exists in the draft
	for i, m := range player.DraftMovies {
		if m.Id == movie.Id {
			player.DraftMovies = append(
				player.DraftMovies[:i],
				player.DraftMovies[i+1:]...,
			)
			action = "deselected"
			wasToggled = true

			rs.logger.Debug("Movie toggled off in draft", "roomName", roomName, "player", username, "movie", movie.Name)
			break
		}
	}

	// Movie not found in draft, try to add it if under limit
	if !wasToggled && len(player.DraftMovies) < room.Game.MaxDraftCount {
		player.DraftMovies = append(player.DraftMovies, movie)
		action = "selected"
		wasToggled = true

		rs.logger.Debug("Movie toggled on in draft", "roomName", roomName, "player", username, "movie", movie)
	}

	if wasToggled && rs.queries != nil {
		go rs.recordVoteEvent(username, "draft_toggle", action, movie)
	}

	return wasToggled
}

func (rs *Service) ToggleVotingMovie(roomName, username string, movie movie.Movie) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	player, ok := room.Players[username]
	if !ok {
		return false
	}

	var wasToggled bool
	var action string

	// Try to remove the movie if it exists in voting
	for i, m := range player.VotingMovies {
		if m.Id == movie.Id {
			player.VotingMovies = append(
				player.VotingMovies[:i],
				player.VotingMovies[i+1:]...,
			)
			action = "deselected"
			wasToggled = true
			rs.logger.Debug("Movie toggled off in Voting", "roomName", roomName, "player", username, "movie", movie.Name)
			break
		}
	}

	// Movie not in voting list, add it
	if !wasToggled {
		player.VotingMovies = append(player.VotingMovies, movie)
		action = "selected"
		wasToggled = true

		rs.logger.Debug("Movie toggled on in Voting", "roomName", roomName, "player",
			username, "movie", movie)
	}

	if wasToggled && rs.queries != nil {
		go rs.recordVoteEvent(username, "vote_toggle", action, movie)
	}

	return wasToggled
}

func (r *Room) GetPlayer(username string) (*Player, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	player, exists := r.Players[username]
	return player, exists
}

func (r *Room) GetAllPlayers() []*Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	players := make([]*Player, 0, len(r.Players))
	for _, player := range r.Players {
		players = append(players, player)
	}
	return players
}

func (r *Room) PlayersByJoinTime() []*Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	players := make([]*Player, 0, len(r.Players))
	for _, u := range r.Players {
		players = append(players, u)
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].JoinedAt.Before(players[j].JoinedAt)
	})
	return players
}

// IsVotingFinished determines if all players have voted on movies, if any have not return false
func (r *Room) IsVotingFinished() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.Players {
		if !p.HasFinishedVoting {
			return false
		}
	}
	return true
}

// IsDraftFinished determines if all players have selected movies, if any have not return false
func (r *Room) IsDraftFinished() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.Players {
		if !p.HasFinishedDraft {
			return false
		}
	}
	return true
}

func (rs *Service) recordVoteEvent(username, eventType, action string,
	movie movie.Movie) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	user, err := rs.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if err != sql.ErrNoRows {
			rs.logger.Error("Failed to get user for event recording", "error", err, "username",
				username)
		}
		return
	}

	_, err = rs.queries.CreateVoteEvent(ctx, sqlcgen.CreateVoteEventParams{
		UserID:    user.ID,
		EventType: eventType,
		Action:    action,
		MovieID:   movie.Id,
		MovieName: movie.Name,
	})

	if err != nil {
		rs.logger.Error("Failed to record vote event",
			"error", err,
			"user", username,
			"type", eventType,
			"action", action,
			"movie", movie.Name)
	} else {
		rs.logger.Debug("Vote event recorded",
			"user", username,
			"type", eventType,
			"action", action,
			"movie", movie.Name)
	}
}

func (rs *Service) SaveGameResult(roomName string) error {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return fmt.Errorf("room not found: %s", roomName)
	}

	if room.Game.Step != Results {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	var winningMovie *movie.Movie
	var winningVoteCount int

	for movie, voteCount := range room.Game.Votes {
		if voteCount > winningVoteCount {
			winningMovie = movie
			winningVoteCount = voteCount
		}
	}

	if winningMovie == nil {
		rs.logger.Warn("Could not determine winning movie", "room", roomName)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gameResult, err := rs.queries.CreateGameResult(ctx, sqlcgen.CreateGameResultParams{
		RoomName:         roomName,
		WinningMovieID:   winningMovie.Id,
		WinningMovieName: winningMovie.Name,
		WinningVoteCount: int64(winningVoteCount),
		TotalPlayers:     int64(len(room.Players)),
	})

	if err != nil {
		rs.logger.Error("Failed to create game result", "error", err, "room", roomName)
		return fmt.Errorf("create game result: %w", err)
	}

	rs.logger.Info("Game result saved",
		"room", roomName,
		"winner", winningMovie.Name,
		"votes", winningVoteCount,
		"players", len(room.Players))

	// Save participants
	for _, player := range room.Players {
		// Get user ID from username
		user, err := rs.queries.GetUserByUsername(ctx, player.Username)
		if err != nil {
			rs.logger.Error("Failed to get user for participant",
				"error", err,
				"username", player.Username)
			continue
		}

		_, err = rs.queries.CreateGameParticipant(ctx, sqlcgen.CreateGameParticipantParams{
			GameID: gameResult.ID,
			UserID: user.ID,
		})

		if err != nil {
			rs.logger.Error("Failed to create game participant",
				"error", err,
				"username", player.Username)
		}
	}

	return nil

}

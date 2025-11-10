package room

import (
	"log/slog"
	"sort"
	"sync"
	"time"

	"watchma/pkg/movie"
)

// Service represents the orchestrator of all rooms and room operations
type Service struct {
	mu     sync.RWMutex
	Rooms  map[string]*Room
	pub    *EventPublisher
	logger *slog.Logger
}

func NewService(pub *EventPublisher, l *slog.Logger) *Service {
	return &Service{
		Rooms:  make(map[string]*Room),
		pub:    pub,
		logger: l,
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

	rs.logger.Info("Player added to room", "roomName", roomName, "playerName", username)

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

	rs.logger.Info("Player removed from room", "roomName", roomName, "playerName", username)

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

func (rs *Service) StreamAnnouncement(roomName string) {
	rs.pub.PublishRoomEvent(roomName, RoomAnnounceEvent)
}

func (rs *Service) FinishGame(roomName string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.Game.Step = Results
	rs.logger.Info("Game Finished", "roomName", roomName)

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

	rs.logger.Info("All Draft Votes Submitted to Voting Array", "Room Name", room.Name)
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

	rs.logger.Info("All Movie Votes submitted to Voting Movies Results Array", "Room Name", room.Name, "votes", room.Game.Votes)
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
			rs.pub.PublishRoomEvent(roomName, RoomUpdateEvent)
			return true
		}
	}

	return false
}

// ToggleDraftMovie adds or removes a movie from a player's draft selection
// If the movie is already in the draft, it will be removed
// If the movie is not in the draft and the player hasn't reached MaxDraftCount, it will be added
// Returns true if the operation was successful
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

	// Try to remove the movie if it exists in the draft
	for i, m := range player.DraftMovies {
		if m.Id == movie.Id {
			player.DraftMovies = append(
				player.DraftMovies[:i],
				player.DraftMovies[i+1:]...,
			)
			rs.logger.Debug("Movie toggled off in draft", "roomName", roomName, "player", username, "movie", movie.Name)
			return true
		}
	}

	// Movie not found in draft, try to add it if under limit
	if len(player.DraftMovies) < room.Game.MaxDraftCount {
		player.DraftMovies = append(player.DraftMovies, movie)
		rs.logger.Debug("Movie toggled on in draft", "roomName", roomName, "player", username, "movie", movie)
		return true
	}

	return false
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

	// Try to remove the movie if it exists in the draft
	for i, m := range player.VotingMovies {
		if m.Id == movie.Id {
			player.VotingMovies = append(
				player.VotingMovies[:i],
				player.VotingMovies[i+1:]...,
			)
			rs.logger.Debug("Movie toggled off in Voting", "roomName", roomName, "player", username, "movie", movie.Name)
			return true
		}
	}

	// Movie not found in draft, try to add it if under limit
	if len(player.VotingMovies) < room.Game.MaxVotes {
		player.VotingMovies = append(player.VotingMovies, movie)
		rs.logger.Debug("Movie toggled on in Voting", "roomName", roomName, "player", username, "movie", movie)
		return true
	}

	return false
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

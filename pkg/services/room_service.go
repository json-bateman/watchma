package services

import (
	"log/slog"
	"sort"
	"sync"
	"time"

	"watchma/pkg/types"
	"watchma/pkg/utils"
)

// All mutation and publish events are done by the RoomService
type RoomService struct {
	mu     sync.RWMutex
	Rooms  map[string]*Room
	pub    *EventPublisher
	logger *slog.Logger
}

// Room is a PURE DATA STRUCTURE (no mutation methods)
type Room struct {
	Name         string
	Game         *types.GameSession
	RoomMessages []types.Message
	Players      map[string]*types.Player
	mu           sync.RWMutex
}

func NewRoomService(pub *EventPublisher, l *slog.Logger) *RoomService {
	return &RoomService{
		Rooms:  make(map[string]*Room),
		pub:    pub,
		logger: l,
	}
}

func (rs *RoomService) AddRoom(roomName string, game *types.GameSession) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Rooms[roomName] = &Room{
		Name:         roomName,
		Game:         game,
		RoomMessages: make([]types.Message, 0),
		Players:      make(map[string]*types.Player),
	}

	rs.logger.Info("Room added", "name", roomName)

	rs.pub.PublishLobbyEvent(utils.ROOM_LIST_UPDATE_EVENT)
}

func (rs *RoomService) DeleteRoom(roomName string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	delete(rs.Rooms, roomName)

	rs.logger.Info("Room deleted", "name", roomName)

	rs.pub.PublishLobbyEvent(utils.ROOM_LIST_UPDATE_EVENT)
}

func (rs *RoomService) AddPlayerToRoom(roomName, username string) (*types.Player, bool) {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return nil, false
	}

	room.mu.Lock()
	defer room.mu.Unlock()
	player := &types.Player{
		Username: username,
		JoinedAt: time.Now(),
	}
	room.Players[username] = player

	rs.logger.Info("Player added to room", "roomName", roomName, "playerName", username)

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	rs.pub.PublishLobbyEvent(utils.ROOM_LIST_UPDATE_EVENT)

	return player, true
}

func (rs *RoomService) RemovePlayerFromRoom(roomName, username string) bool {
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

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	rs.pub.PublishLobbyEvent(utils.ROOM_LIST_UPDATE_EVENT)

	return true
}

func (rs *RoomService) TransferHost(roomName, newHost string) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	room.Game.Host = newHost
	room.mu.Unlock()

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	return true
}

func (rs *RoomService) TogglePlayerReady(roomName, username string) bool {
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

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	return true
}

func (rs *RoomService) TogglePlayerFinishedDraft(roomName, username string) bool {
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
	player.HasFinishedDraft = !player.HasFinishedDraft

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_UPDATE_EVENT)
	return true
}

func (rs *RoomService) AddMessage(roomName string, msg types.Message) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	defer room.mu.Unlock()
	room.RoomMessages = append(room.RoomMessages, msg)

	rs.pub.PublishRoomEvent(roomName, utils.MESSAGE_SENT_EVENT)
	return true
}

func (rs *RoomService) StartGame(roomName string, movies []types.Movie) bool {
	room, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	room.mu.Lock()
	defer room.mu.Unlock()
	room.Game.Step = types.Voting
	room.Game.Movies = movies

	rs.logger.Info("Game started", "roomName", roomName)

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_START_EVENT)
	return true
}

func (rs *RoomService) FinishGame(roomName string) bool {
	_, ok := rs.GetRoom(roomName)
	if !ok {
		return false
	}

	rs.logger.Info("Game finished", "roomName", roomName)

	rs.pub.PublishRoomEvent(roomName, utils.ROOM_FINISH_EVENT)
	return true
}

func (rs *RoomService) RoomExists(roomName string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	_, exists := rs.Rooms[roomName]
	return exists
}

func (rs *RoomService) GetRoom(roomName string) (*Room, bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	room, ok := rs.Rooms[roomName]
	return room, ok
}

// Submits movie votes for a given user and returns whether the voting period is concluded
// for the given room. When voting is concluded, we can advance the step
func (rs *RoomService) SubmitVotes(roomName string, username string, movies []string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	room, ok := rs.GetRoom(roomName)
	player, ok2 := room.GetPlayer(username)

	if ok && ok2 {
		for _, movieID := range movies {
			// Find the Movie that matches this ID
			for i := range room.Game.Movies {
				if room.Game.Movies[i].Id == movieID {
					room.Game.Votes[&room.Game.Movies[i]]++
					break
				}
			}
			player.SelectedMovies = append(player.SelectedMovies, movieID)
		}
	}

	rs.logger.Info("Player submitted votes", "roomName", roomName, "player", username, "votes", movies)

	player.HasSelectedMovies = true

	return rs.GetIsVotingFinished(roomName)
}

func (rs *RoomService) GetIsVotingFinished(roomName string) bool {
	room, _ := rs.GetRoom(roomName)

	for _, player := range room.Players {
		if !player.HasSelectedMovies {
			return false
		}
	}

	return true
}

func (r *Room) GetPlayer(username string) (*types.Player, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	player, exists := r.Players[username]
	return player, exists
}

func (r *Room) GetAllPlayers() []*types.Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	players := make([]*types.Player, 0, len(r.Players))
	for _, player := range r.Players {
		players = append(players, player)
	}
	return players
}

func (r *Room) PlayersByJoinTime() []*types.Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	players := make([]*types.Player, 0, len(r.Players))
	for _, u := range r.Players {
		players = append(players, u)
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].JoinedAt.Before(players[j].JoinedAt)
	})
	return players
}

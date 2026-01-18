package game

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"watchma/db/sqlcgen"
	appctx "watchma/pkg/context"
	"watchma/pkg/movie"
	"watchma/pkg/openai"
	"watchma/pkg/room"
	"watchma/web"
	"watchma/web/features/game/pages"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"
)

type handlers struct {
	roomService    *room.Service
	movieService   *movie.Service
	openAiProvider *openai.Provider
	logger         *slog.Logger
	nats           *nats.Conn
}

func newHandlers(
	roomService *room.Service,
	movieService *movie.Service,
	openAiProvider *openai.Provider,
	logger *slog.Logger,
	nc *nats.Conn,
) *handlers {
	return &handlers{
		roomService:    roomService,
		movieService:   movieService,
		openAiProvider: openAiProvider,
		logger:         logger,
		nats:           nc,
	}
}

// ============= LOBBY HANDLERS =============

func (h *handlers) singleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user, ok := h.getUserFromRequest(w, r)
	if !ok {
		return
	}

	myRoom, ok := h.getRoomByName(w, r, roomName)
	if !ok {
		return
	}

	// Check if game has already started
	if myRoom.Game.Step != room.Lobby {
		// Game in progress - check if player is already in the room
		_, playerInRoom := myRoom.GetPlayer(user.Username)
		if !playerInRoom {
			// Player not in room and game started - redirect to home
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		// Player is in room - let them reconnect (fall through)
	}

	if myRoom.Game.MaxPlayers <= len(myRoom.Players) {
		web.RenderPage(pages.RoomFull(), roomName, w, r)
		return
	}

	h.roomService.AddPlayerToRoom(myRoom.Name, user.Username)

	web.RenderPageNoLayout(pages.Lobby(myRoom, user.Username), myRoom.Name, w, r)
}

// Function that does the heavy lifting by keeping the SSE channel open and sending
// Events to the client in real-time
func (h *handlers) singleRoomSSE(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := appctx.GetUserFromRequest(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sse := datastar.NewSSE(w, r)

	// Check if room exists
	myRoom, ok := h.roomService.GetRoom(roomName)
	if !ok {
		h.logger.Warn("Room not found on SSE reconnect", "Room", roomName, "Username", user.Username)
		if err := sse.Redirect("/"); err != nil {
			h.logger.Warn("Error redirecting after room not found", "error", err)
		}
		return
	}

	// Send existing user list to new client
	userBox := pages.UserBox(myRoom, user.Username)
	if err := sse.PatchElementTempl(userBox); err != nil {
		h.logger.Error("Error patching initial user list", "error", err)
	}

	// Send existing messages to new client
	if len(myRoom.RoomMessages) > 0 {
		chat := pages.ChatBox(myRoom.RoomMessages)
		if err := sse.PatchElementTempl(chat); err != nil {
			h.logger.Error("Error patching chatbox on load", "error", err)
			return
		}
	}

	// Subscribe to room-specific NATS subject
	roomSubject := room.RoomSubject(roomName)
	sub, err := h.nats.SubscribeSync(roomSubject)
	h.logger.Debug(room.NATSSub, "subject", roomSubject)
	if err != nil {
		http.Error(w, "Subscribe Failed", http.StatusInternalServerError)
		return
	}

	for {
		msg, err := sub.NextMsgWithContext(r.Context())
		if err != nil {
			// context canceled or sub closed
			return
		}
		switch string(msg.Data) {
		case room.RoomUpdateEvent:
			userBox := pages.UserBox(myRoom, user.Username)
			if err := sse.PatchElementTempl(userBox); err != nil {
				h.logger.Error("Error patching user list", "error", err)
				return
			}
		case room.MessageSentEvent:
			chat := pages.ChatBox(myRoom.RoomMessages)
			if err := sse.PatchElementTempl(chat); err != nil {
				h.logger.Error("Error patching chat message", "error", err)
				return
			}
		case room.RoomStartEvent:
			player, ok := h.getPlayerInRoom(myRoom, user.Username)
			if !ok {
				return
			}
			draftPage := pages.Draft(player, myRoom)
			if err := sse.PatchElementTempl(draftPage); err != nil {
				h.logger.Error("Error patching draft page", "error", err)
				return
			}
		case room.RoomVotingEvent:
			player, ok := h.getPlayerInRoom(myRoom, user.Username)
			if !ok {
				return
			}
			votingPage := pages.Voting(myRoom.Game.VotingMovies, player, myRoom)
			if err := sse.PatchElementTempl(votingPage); err != nil {
				h.logger.Error("Error patching voting page", "error", err)
				return
			}
		case room.RoomAnnounceEvent:
			streamedMessagePage := pages.AiAnnounce(myRoom, myRoom.Game.Announcement)
			if err := sse.PatchElementTempl(streamedMessagePage); err != nil {
				return
			}
		case room.RoomFinishEvent:
			movieVotes := sortMoviesByVotes(myRoom.Game.Votes)
			winnerMovies := getWinnerMovies(movieVotes)
			resultsPage := pages.ResultsScreen(winnerMovies, myRoom)
			if err := sse.PatchElementTempl(resultsPage); err != nil {
				h.logger.Error("Error patching results page", "error", err)
				return
			}

		default: // discard for now, maybe error?
		}
	}
}

func (h *handlers) leaveRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user, ok := h.getUserFromRequest(w, r)
	if !ok {
		return
	}

	myRoom, ok := h.getRoomByName(w, r, roomName)
	if !ok {
		return
	}

	h.roomService.RemovePlayerFromRoom(myRoom.Name, user.Username)

	allUsers := myRoom.GetAllPlayers()
	if len(allUsers) == 0 {
		h.roomService.DeleteRoom(myRoom.Name)
		return
	}

	if myRoom.Game.Host == user.Username {
		// If host leaves transfer to random other user
		for newHostUsername := range myRoom.Players {
			h.roomService.TransferHost(myRoom.Name, newHostUsername)
			break
		}
	}
}

func (h *handlers) publishChatMessage(w http.ResponseWriter, r *http.Request) {
	var req room.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}
	if len(strings.Trim(req.Message, " ")) == 0 {
		return
	}

	user, ok := h.getUserFromRequest(w, r)
	if !ok {
		return
	}

	req.Username = user.Username
	myRoom, ok := h.roomService.GetRoom(req.Room)
	if ok {
		h.roomService.AddMessage(myRoom.Name, req)
	}
}

func (h *handlers) ready(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user, ok := h.getUserFromRequest(w, r)
	if !ok {
		return
	}

	myRoom, ok := h.roomService.GetRoom(roomName)
	if ok {
		h.roomService.TogglePlayerReady(myRoom.Name, user.Username)
	}
}

func (h *handlers) startGame(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	myRoom, ok := h.roomService.GetRoom(roomName)
	if ok {
		myRoom.Game.Step = room.Draft
		movies, err := h.movieService.GetMovies()
		if err != nil {
			h.logger.Error("Call to MovieService.GetMovies failed", "Error", err)
			return
		}

		if len(movies) == 0 {
			h.logger.Warn(fmt.Sprintf("Room %s: No Movies Found", myRoom.Name))
		}

		myRoom.Game.AllMovies = movies

		h.roomService.StartGame(roomName, myRoom.Game.AllMovies)
	}
}

// ============= DRAFT HANDLERS =============

type movieQueryRequest struct {
	Search string `json:"search"`
	Genre  string `json:"genre"`
	Sort   string `json:"sort"`
}

func (h *handlers) draft(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	myRoom, _, player, ok := h.getRoomUserAndPlayer(w, r, roomName)
	if !ok {
		return
	}

	web.RenderPageNoLayout(pages.Draft(player, myRoom), myRoom.Name, w, r)
}

func (h *handlers) deleteFromSelectedMovies(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	movieId := chi.URLParam(r, "id")
	_, _, player, ok := h.getRoomUserAndPlayer(w, r, roomName)
	if !ok {
		return
	}

	// Prevent changes if already submitted
	if player.HasFinishedDraft {
		h.logger.Warn("Cannot modify draft after submission", "Room", roomName, "Username", player.Username)
		h.renderDraftPage(w, r)
		return
	}

	// Use RoomService to handle business logic
	if !h.roomService.RemoveDraftMovie(roomName, player.Username, movieId) {
		h.logger.Warn("Failed to remove draft movie", "Room", roomName, "Username", player.Username, "MovieId", movieId)
	}

	h.renderDraftPage(w, r)
}

func (h *handlers) toggleDraftMovie(w http.ResponseWriter, r *http.Request) {
	movieId := chi.URLParam(r, "id")
	roomName := chi.URLParam(r, "roomName")
	_, user, player, ok := h.getRoomUserAndPlayer(w, r, roomName)
	if !ok {
		return
	}

	// Prevent changes if already submitted
	if player.HasFinishedDraft {
		h.logger.Warn("Cannot modify draft after submission", "Room", roomName, "Username", user.Username)
		h.renderDraftPage(w, r)
		return
	}

	var mov movie.Movie

	// Use the player's own copy of available movies
	for _, m := range player.AvailableMovies {
		if m.Id == movieId {
			mov = m
			break
		}
	}

	if !h.roomService.ToggleDraftMovie(roomName, user.Username, mov) {
		h.logger.Warn("Failed to toggle draft movie", "Room", roomName, "Username", user.Username, "MovieId", movieId)
	}

	h.renderDraftPage(w, r)
}

func (h *handlers) queryMovies(w http.ResponseWriter, r *http.Request) {
	h.renderDraftPage(w, r)
}

func (h *handlers) draftSubmit(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	myRoom, _, player, ok := h.getRoomUserAndPlayer(w, r, roomName)
	if !ok {
		return
	}

	if len(player.DraftMovies) == 0 {
		web.SendSSEError(w, r, "Must include at least 1 movie id.", h.logger)
		return
	}

	player.HasFinishedDraft = true

	isDraftFinished := myRoom.IsDraftFinished()

	// if voting is finished, add all players choices to the voting array
	if isDraftFinished {
		myRoom.Game.Step = room.Voting
		h.roomService.SubmitDraftVotes(myRoom)
		h.roomService.MoveToVoting(roomName)
	} else {
		h.renderDraftPage(w, r)
	}
}

func (h *handlers) renderDraftPage(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	myRoom, _, player, ok := h.getRoomUserAndPlayer(w, r, roomName)
	if !ok {
		return
	}

	var queryRequest movieQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&queryRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		h.logger.Warn("Error decoding SortFilter", "error", err)
		return
	}

	var sortField movie.SortField
	descending := false
	switch queryRequest.Sort {
	case "name-asc":
		sortField = movie.SortByName
	case "name-desc":
		sortField = movie.SortByName
		descending = true
	case "year-asc":
		sortField = movie.SortByYear
	case "year-desc":
		sortField = movie.SortByYear
		descending = true
	case "critic-asc":
		sortField = movie.SortByCriticRating
	case "critic-desc":
		sortField = movie.SortByCriticRating
		descending = true
	case "community-asc":
		sortField = movie.SortByCommunityRating
	case "community-desc":
		sortField = movie.SortByCommunityRating
		descending = true
	default:
		// default to name instead of blowing up for now
		// sortField = movie.SortByName
	}

	var err error
	player.AvailableMovies, err = h.movieService.GetMoviesWithQuery(
		movie.Query{
			Search:     queryRequest.Search,
			Genre:      queryRequest.Genre,
			SortBy:     sortField,
			Descending: descending,
		},
	)

	if err != nil {
		h.logger.Error("Movie Query Error", "Error", err)
	}
	draft := pages.Draft(player, myRoom)
	if err := datastar.NewSSE(w, r).PatchElementTempl(draft); err != nil {
		h.logger.Error("Error Rendering Draft Page", "error", err)
	}
}

// ============= VOTING HANDLERS =============

func (h *handlers) voting(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	myRoom, _, player, ok := h.getRoomUserAndPlayer(w, r, roomName)
	if !ok {
		return
	}

	web.RenderPageNoLayout(pages.Voting(myRoom.Game.VotingMovies, player, myRoom), myRoom.Name, w, r)
}

func (h *handlers) votingSubmit(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	myRoom, _, player, ok := h.getRoomUserAndPlayer(w, r, roomName)
	if !ok {
		return
	}

	if len(player.VotingMovies) == 0 {
		web.SendSSEError(w, r, "Must include at least 1 movie id.", h.logger)
		return
	}

	player.HasFinishedVoting = true

	isVotingFinished := myRoom.IsVotingFinished()

	if isVotingFinished {
		h.roomService.SubmitFinalVotes(myRoom)

		movieVotes := sortMoviesByVotes(myRoom.Game.Votes)
		tiedMovies := getWinnerMovies(movieVotes)
		myRoom.Game.VotingNumber = tiedMovies[0].Votes
		if len(tiedMovies) > 1 {
			// If there's a tie, find number of votes and reset players
			// so they can vote again
			tied := make([]movie.Movie, 0, len(tiedMovies))
			for _, m := range tiedMovies {
				tied = append(tied, *m.Movie)
			}
			// reset vote map for game and voting movies for each player
			myRoom.Game.Votes = map[*movie.Movie]int{}
			myRoom.Game.VotingMovies = tied
			for _, p := range myRoom.Players {
				p.VotingMovies = []movie.Movie{}
				p.HasFinishedVoting = false
			}
			h.logger.Info("Tie detected, moving to revote", "roomName", roomName, "tiedMovies", len(tiedMovies), "votes", tiedMovies[0].Votes)
			h.roomService.MoveToVoting(myRoom.Name)
		} else {
			myRoom.Game.Step = room.Announce
			// Trigger announcement with AI generation
			h.generateAndStreamAnnouncement(roomName, tiedMovies[0].Movie)
			// h.roomService.AnnounceWinner(roomName)
		}
	} else {
		h.renderVotingPage(w, r)
	}
}

func (h *handlers) toggleVotingMovie(w http.ResponseWriter, r *http.Request) {
	movieId := chi.URLParam(r, "id")
	roomName := chi.URLParam(r, "roomName")
	myRoom, user, player, ok := h.getRoomUserAndPlayer(w, r, roomName)
	if !ok {
		return
	}

	// Prevent changes if already submitted
	if player.HasFinishedVoting {
		h.logger.Warn("Cannot modify votes after submission", "Room", roomName, "Username", user.Username)
		h.renderVotingPage(w, r)
		return
	}

	movie := myRoom.Game.AllMoviesMap[movieId]

	if !h.roomService.ToggleVotingMovie(roomName, user.Username, *movie) {
		h.logger.Warn("Failed to toggle steps.Voting movie", "Room", roomName, "Username", user.Username, "MovieId", movieId)
	}

	h.renderVotingPage(w, r)
}

func (h *handlers) renderVotingPage(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	myRoom, _, player, ok := h.getRoomUserAndPlayer(w, r, roomName)
	if !ok {
		return
	}

	draft := pages.Voting(myRoom.Game.VotingMovies, player, myRoom)
	if err := datastar.NewSSE(w, r).PatchElementTempl(draft); err != nil {
		h.logger.Error("Error Rendering Voting Page", "error", err)
	}
}

// ============= RESULTS HANDLERS =============

func (h *handlers) results(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	myRoom, _, _, ok := h.getRoomUserAndPlayer(w, r, roomName)
	if !ok {
		return
	}

	movieVotes := sortMoviesByVotes(myRoom.Game.Votes)
	winnerMovies := getWinnerMovies(movieVotes)

	web.RenderPage(pages.ResultsScreen(winnerMovies, myRoom), myRoom.Name, w, r)
}

// =============== HELPERS ================

// sortMoviesByVotes converts a vote map to a sorted slice (descending by votes)
func sortMoviesByVotes(votes map[*movie.Movie]int) []movie.Vote {
	movieVotes := make([]movie.Vote, 0, len(votes))

	for m, voteCount := range votes {
		movieVotes = append(movieVotes, movie.Vote{
			Movie: m,
			Votes: voteCount,
		})
	}

	sort.Slice(movieVotes, func(i, j int) bool {
		return movieVotes[i].Votes > movieVotes[j].Votes
	})

	return movieVotes
}

// getWinnerMovies returns only movies with the highest vote count
// Since sortMoviesByVotes is ranging over a map, the selection is random
func getWinnerMovies(moviesSortedByVote []movie.Vote) []movie.Vote {
	if len(moviesSortedByVote) == 0 {
		return []movie.Vote{}
	}

	maxVotes := moviesSortedByVote[0].Votes
	var winners []movie.Vote

	for _, movie := range moviesSortedByVote {
		if movie.Votes == maxVotes {
			winners = append(winners, movie)
		}
	}

	return winners
}

// generateAndStreamAnnouncement runs AI generation once and streams to all clients via NATS
func (h *handlers) generateAndStreamAnnouncement(roomName string, winnerMovie *movie.Movie) {
	myRoom, ok := h.roomService.GetRoom(roomName)
	if !ok {
		return
	}

	// Initial drum roll
	myRoom.Game.Announcement = []room.DialogueLine{{
		Character: "Announcer ",
		Dialogue:  "Drum Roll Please",
	}}
	h.roomService.StreamAnnouncement(roomName)
	time.Sleep(2 * time.Second)

	// Clear and generate AI dialogue
	myRoom.Game.Announcement = []room.DialogueLine{}

	buildGptMessage := fmt.Sprintf(`You are writing a reveal scene for: %s

  Write a dialogue-only scene using this EXACT format:

  **[Character Name]:** *"their dialogue"*
  **[Different Character]:** *"their response"*

  RULES:
  1. Use 2-4 characters from the movie
  2. Each character speaks 1-2 times
  3. Include character catchphrases naturally
  4. Build suspense without saying the movie title
  5. NO actor names, NO spoilers, NO narration, NO movie title
  6. Match the movie's genre/tone
  7. Keep your responses TERSE

  EXAMPLE (for a different movie):
  **Morpheus:** *"What if I told you... we're the ones they chose?"*
  **Trinity:** *"The question isn't how, it's why."*

  NOW write the scene for:`, winnerMovie.Name)

	var finalMessage string
	if h.openAiProvider != nil {
		var err error
		finalMessage, err = h.openAiProvider.FetchAiResponse(buildGptMessage)
		if err != nil {
			h.logger.Error("AI request failed", "error", err)
		}
	}

	dialogueRegex := regexp.MustCompile(`\*\*\[?([^\]:]+)\]?:\*\*\s*\*"([^"]+)"\*`)
	matches := dialogueRegex.FindAllStringSubmatch(finalMessage, -1)

	// Stream each dialogue line
	for _, match := range matches {
		if len(match) == 3 {
			line := room.DialogueLine{
				Character: strings.TrimSpace(match[1]),
				Dialogue:  strings.TrimSpace(match[2]),
			}
			myRoom.Game.Announcement = append(myRoom.Game.Announcement, line)
			h.roomService.StreamAnnouncement(roomName)
			time.Sleep(2 * time.Second)
		}
	}

	// Final announcement
	myRoom.Game.Announcement = []room.DialogueLine{{
		Character: "Announcer",
		Dialogue:  "And the Winner Is...",
	}}
	h.roomService.StreamAnnouncement(roomName)
	time.Sleep(2 * time.Second)

	h.roomService.FinishGame(roomName)
}

// =============== VALIDATION HELPERS ================

// getUserFromRequest retrieves and validates the user from the request context
func (h *handlers) getUserFromRequest(w http.ResponseWriter, r *http.Request) (*sqlcgen.User, bool) {
	user := appctx.GetUserFromRequest(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	return user, true
}

// getRoomByName retrieves a room by name and handles error response if not found
func (h *handlers) getRoomByName(w http.ResponseWriter, r *http.Request, roomName string) (*room.Room, bool) {
	myRoom, ok := h.roomService.GetRoom(roomName)
	if !ok {
		web.RenderPage(pages.NoRoom(roomName), roomName, w, r)
		return nil, false
	}
	return myRoom, true
}

// getPlayerInRoom retrieves a player from a room and logs if not found
func (h *handlers) getPlayerInRoom(myRoom *room.Room, username string) (*room.Player, bool) {
	player, ok := myRoom.GetPlayer(username)
	if !ok {
		h.logger.Warn("Player not in room", "username", username, "room", myRoom.Name)
		return nil, false
	}
	return player, true
}

// getRoomUserAndPlayer combines all three validation steps for convenience
func (h *handlers) getRoomUserAndPlayer(w http.ResponseWriter, r *http.Request, roomName string) (*room.Room, *sqlcgen.User, *room.Player, bool) {
	user, ok := h.getUserFromRequest(w, r)
	if !ok {
		return nil, nil, nil, false
	}

	myRoom, ok := h.getRoomByName(w, r, roomName)
	if !ok {
		return nil, nil, nil, false
	}

	player, ok := h.getPlayerInRoom(myRoom, user.Username)
	if !ok {
		http.Error(w, "Player not in room", http.StatusUnauthorized)
		return nil, nil, nil, false
	}

	return myRoom, user, player, true
}

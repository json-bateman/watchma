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
	user := appctx.GetUserFromRequest(r)

	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	myRoom, ok := h.roomService.GetRoom(roomName)

	if !ok {
		web.RenderPage(pages.NoRoom(roomName), roomName, w, r)
		return
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
		// Send error and redirect to home
		if err := sse.ExecuteScript("window.location.href = '/'"); err != nil {
			h.logger.Warn("Error redirecting after room not found", "error", err)
		}
		return
	}

	// Send existing user list to new client
	userBox := pages.UserBox(myRoom, user.Username)
	if err := sse.PatchElementTempl(userBox); err != nil {
		h.logger.Error("Error patching initial user list", "error", err)
	}

	player, ok := myRoom.GetPlayer(user.Username)
	if !ok {
		h.logger.Warn("User not in room", "Username", user.Username, "Room", myRoom.Name)
		// Send error and redirect to home
		if err := sse.ExecuteScript("window.location.href = '/'"); err != nil {
			h.logger.Warn("Error redirecting after player not found", "error", err)
		}
		return
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

	movies, err := h.movieService.GetMovies()
	if err != nil {
		h.logger.Error("Call to MovieService.GetMovies failed", "Error", err)
		return
	}

	defer func() {
		sub.Unsubscribe()
		h.leaveRoom(w, r)
	}()

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
			draft := pages.Draft(player, movies, myRoom)
			if err := sse.PatchElementTempl(draft); err != nil {
				h.logger.Error("Error patching movies", "error", err)
				return
			}
		case room.RoomVotingEvent:
			movies := pages.Voting(myRoom.Game.VotingMovies, player, myRoom)
			if err := sse.PatchElementTempl(movies); err != nil {
				h.logger.Error("Error patching movies", "error", err)
				return
			}
		case room.RoomAnnounceEvent:
			movies := pages.AiAnnounce(myRoom, myRoom.Game.Announcement)
			if err := sse.PatchElementTempl(movies); err != nil {
				h.logger.Error("Error patching movies", "error", err)
				return
			}
		case room.RoomFinishEvent:
			movieVotes := sortMoviesByVotes(myRoom.Game.Votes)
			winnerMovies := getWinnerMovies(movieVotes)
			finalScreen := pages.ResultsScreen(winnerMovies)
			if err := sse.PatchElementTempl(finalScreen); err != nil {
				h.logger.Error("Error patching final screen", "error", err)
				return
			}

		default: // discard for now, maybe error?
		}
	}
}

func (h *handlers) leaveRoom(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := appctx.GetUserFromRequest(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := h.roomService.GetRoom(roomName)
	if !ok {
		web.WriteJSONError(w, http.StatusNotFound, "Room not found")
		return
	}

	h.roomService.RemovePlayerFromRoom(room.Name, user.Username)

	allUsers := room.GetAllPlayers()
	if len(allUsers) == 0 {
		h.roomService.DeleteRoom(room.Name)
		return
	}

	if room.Game.Host == user.Username {
		// If host leaves transfer to random other user
		for newHostUsername := range room.Players {
			h.roomService.TransferHost(room.Name, newHostUsername)
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

	user := appctx.GetUserFromRequest(r)
	if user == nil {
		web.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	req.Username = user.Username
	room, ok := h.roomService.GetRoom(req.Room)
	if ok {
		h.roomService.AddMessage(room.Name, req)
	}
}

func (h *handlers) ready(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	user := appctx.GetUserFromRequest(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := h.roomService.GetRoom(roomName)
	if ok {
		h.roomService.TogglePlayerReady(room.Name, user.Username)
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

func (h *handlers) deleteFromSelectedMovies(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	movieId := chi.URLParam(r, "id")
	user := appctx.GetUserFromRequest(r)

	// Use RoomService to handle business logic
	if !h.roomService.RemoveDraftMovie(roomName, user.Username, movieId) {
		h.logger.Warn("Failed to remove draft movie", "Room", roomName, "Username", user.Username, "MovieId", movieId)
	}

	h.renderDraftPage(w, r)
}

func (h *handlers) toggleDraftMovie(w http.ResponseWriter, r *http.Request) {
	movieId := chi.URLParam(r, "id")
	user := appctx.GetUserFromRequest(r)
	roomName := chi.URLParam(r, "roomName")
	room, _ := h.roomService.GetRoom(roomName)

	var mov movie.Movie

	for _, m := range room.Game.AllMovies {
		if m.Id == movieId {
			mov = m
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
	myRoom, ok := h.roomService.GetRoom(roomName)

	currentUser := appctx.GetUserFromRequest(r)
	if currentUser == nil {
		h.logger.Warn("No User found from session cookie")
		return
	}

	player, ok := myRoom.GetPlayer(currentUser.Username)
	if !ok {
		h.logger.Warn("Player not in room")
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
	room, ok := h.roomService.GetRoom(roomName)

	if !ok {
		h.logger.Warn("Could not obtain room", "room", roomName)
		return
	}

	currentUser := appctx.GetUserFromRequest(r)
	if currentUser == nil {
		h.logger.Warn("No User found from session cookie")
		return
	}

	player, ok := room.GetPlayer(currentUser.Username)
	if !ok {
		h.logger.Warn("Player not in room")
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

	movies, err := h.movieService.GetMoviesWithQuery(
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
	draft := pages.Draft(player, movies, room)
	if err := datastar.NewSSE(w, r).PatchElementTempl(draft); err != nil {
		h.logger.Error("Error Rendering Draft Page", "error", err)
	}
}

// ============= VOTING HANDLERS =============

func (h *handlers) votingSubmit(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	myRoom, ok := h.roomService.GetRoom(roomName)

	currentUser := appctx.GetUserFromRequest(r)
	if currentUser == nil {
		h.logger.Warn("No User found from session cookie")
		return
	}

	player, ok := myRoom.GetPlayer(currentUser.Username)
	if !ok {
		h.logger.Warn("Player not in room")
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
			h.roomService.MoveToVoting(myRoom.Name)
		} else {
			myRoom.Game.Step = room.Results
			h.roomService.AnnounceWinner(roomName)
		}
	} else {
		h.renderVotingPage(w, r)
	}
}

func (h *handlers) toggleVotingMovie(w http.ResponseWriter, r *http.Request) {
	movieId := chi.URLParam(r, "id")
	user := appctx.GetUserFromRequest(r)
	roomName := chi.URLParam(r, "roomName")
	room, _ := h.roomService.GetRoom(roomName)

	movie := room.Game.AllMoviesMap[movieId]

	if !h.roomService.ToggleVotingMovie(roomName, user.Username, *movie) {
		h.logger.Warn("Failed to toggle steps.Voting movie", "Room", roomName, "Username", user.Username, "MovieId", movieId)
	}

	h.renderVotingPage(w, r)
}

func (h *handlers) renderVotingPage(w http.ResponseWriter, r *http.Request) {
	// Voting page needs all movies, and the room
	roomName := chi.URLParam(r, "roomName")
	room, ok := h.roomService.GetRoom(roomName)

	if !ok {
		h.logger.Warn("Could not obtain room", "room", roomName)
		return
	}

	currentUser := appctx.GetUserFromRequest(r)
	if currentUser == nil {
		h.logger.Warn("No User found from session cookie")
		return
	}

	player, ok := room.GetPlayer(currentUser.Username)
	if !ok {
		h.logger.Warn("Player not in room")
		return
	}

	draft := pages.Voting(room.Game.VotingMovies, player, room)
	if err := datastar.NewSSE(w, r).PatchElementTempl(draft); err != nil {
		h.logger.Error("Error Rendering Voting Page", "error", err)
	}
}

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

// ============= ANNOUNCE HANDLERS =============
func (h *handlers) announce(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "roomName")
	myRoom, _ := h.roomService.GetRoom(roomName)
	user := appctx.GetUserFromRequest(r)
	player, ok := myRoom.GetPlayer(user.Username)

	if !ok {
		h.logger.Warn("Player not in room")
		return
	}

	sorted := sortMoviesByVotes(myRoom.Game.Votes)
	winners := getWinnerMovies(sorted)

	if player.Username != myRoom.Game.Host {
		// Exit early, only the host generates the GPT response for the room
		return
	}

	myRoom.Game.Announcement = []room.DialogueLine{
		{
			Character: "Announcer: ",
			Dialogue:  "Drum Roll Please",
		},
	}
	h.roomService.StreamAnnouncement(roomName)
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

  NOW write the scene for:`, winners[0].Movie.Name)

	var finalMessage string
	if h.openAiProvider != nil {
		var err error
		finalMessage, err = h.openAiProvider.FetchAiResponse(buildGptMessage)
		if err != nil {
			h.logger.Error("AI request failed", "error", err)
		}
	}

	dialogueRegex := regexp.MustCompile(`\*\*\[?([^\]:]+)\]?:\*\*\s*\*"([^"]+)"\*`)

	// Example len() == 3 Match:
	// ["**[Trinity]:** *"The question isn't how, it's why."*", "Trinity", "The question isn't how, it's why."]

	matches := dialogueRegex.FindAllStringSubmatch(finalMessage, -1)
	lines := make([]room.DialogueLine, 0, len(matches))

	for _, match := range matches {
		if len(match) == 3 {
			lines = append(lines, room.DialogueLine{
				Character: strings.TrimSpace(match[1]),
				Dialogue:  strings.TrimSpace(match[2]),
			})
		}
	}

	for _, line := range lines {
		myRoom.Game.Announcement = append(myRoom.Game.Announcement, line)
		h.roomService.StreamAnnouncement(roomName)
		time.Sleep(2000 * time.Millisecond)
	}

	time.Sleep(4000 * time.Millisecond)
	myRoom.Game.Announcement = []room.DialogueLine{{
		Character: "Announcer",
		Dialogue:  "And the Winner Is...",
	}}

	h.roomService.StreamAnnouncement(roomName)
	time.Sleep(5000 * time.Millisecond)

	h.roomService.FinishGame(roomName)
}

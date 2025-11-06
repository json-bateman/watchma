package types

// Dumping unorganized structs in this file until there's enough to refactor

type GameStep int

const (
	Lobby = iota
	Draft
	Voting
	Announce
	Results
)

type GameSession struct {
	Host          string
	AllMovies     []Movie
	AllMoviesMap  map[string]*Movie // for fast lookup by ID
	VotingMovies  []Movie
	MaxPlayers    int
	MaxDraftCount int
	MaxVotes      int
	DisplayTies   bool
	Votes         map[*Movie]int // MovieID -> vote count
	Step          GameStep
}

// MovieRequest represents a request containing movie IDs
type MovieRequest struct {
	Movies []string `json:"movies"`
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
	Movie *Movie
	Votes int
}

// Helper methods for GameSession

// SetAllMovies sets the movies and builds the lookup map
func (g *GameSession) SetAllMovies(movies []Movie) {
	g.AllMovies = movies
	g.AllMoviesMap = make(map[string]*Movie, len(movies))
	for i := range movies {
		g.AllMoviesMap[movies[i].Id] = &movies[i]
	}
}

// GetMovie returns a movie by ID in O(1) time
func (g *GameSession) GetMovie(movie Movie) (*Movie, bool) {
	foundMovie, ok := g.AllMoviesMap[movie.Id]
	return foundMovie, ok
}

// VotingMoviesContains checks if a movie ID already exists in VotingMovies
func (g *GameSession) VotingMoviesContains(movie Movie) bool {
	for _, m := range g.VotingMovies {
		if m.Id == movie.Id {
			return true
		}
	}
	return false
}

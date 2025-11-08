package room

import "watchma/pkg/movie"

type Step int

const (
	Lobby Step = iota
	Draft
	Voting
	Announce
	Results
)

type Session struct {
	Host          string
	AllMovies     []movie.Movie
	AllMoviesMap  map[string]*movie.Movie // for fast lookup by ID
	VotingMovies  []movie.Movie
	MaxPlayers    int
	MaxDraftCount int
	MaxVotes      int
	DisplayTies   bool
	Votes         map[*movie.Movie]int // Movie -> vote count
	Step          Step
}

// Helper methods for Session

// SetAllMovies sets the movies and builds the lookup map
func (g *Session) SetAllMovies(movies []movie.Movie) {
	g.AllMovies = movies
	g.AllMoviesMap = make(map[string]*movie.Movie, len(movies))
	for i := range movies {
		g.AllMoviesMap[movies[i].Id] = &movies[i]
	}
}

// GetMovie returns a movie by ID in O(1) time
func (g *Session) GetMovie(m movie.Movie) (*movie.Movie, bool) {
	foundMovie, ok := g.AllMoviesMap[m.Id]
	return foundMovie, ok
}

// VotingMoviesContains checks if a movie ID already exists in VotingMovies
func (g *Session) VotingMoviesContains(m movie.Movie) bool {
	for _, vm := range g.VotingMovies {
		if vm.Id == m.Id {
			return true
		}
	}
	return false
}

// Message represents a chat message
type Message struct {
	Subject  string `json:"subject"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Room     string `json:"room"`
}

// DialogueLine represents the parsing of gippity
type DialogueLine struct {
	Character string
	Dialogue  string
}

package api

type GameStep int

const (
	Lobby GameStep = iota
	Voting
	Results
	Finished
)

type GameSession struct {
	ID     string           // Unique session ID
	Movies []string         // The 20 movies to choose from
	Votes  map[string]int   // Movie -> vote count
	Users  map[string]*User // UserID -> User info
	Step   GameStep         // Game State Step
}

type User struct {
	ID   string
	Name string
}

package api

type GameStep int

const (
	StepLobby GameStep = iota
	StepVoting
	StepResults
	StepFinished
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

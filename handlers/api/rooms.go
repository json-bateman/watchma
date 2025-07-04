package api

func Game() {
}

func (s GameStep) String() string {
	switch s {
	case Lobby:
		return "lobby"
	case Voting:
		return "voting"
	case Results:
		return "results"
	case Finished:
		return "finished"
	default:
		return "unknown"
	}
}

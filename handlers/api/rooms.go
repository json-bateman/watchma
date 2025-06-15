package api

func Game() {
}

func (s GameStep) String() string {
	switch s {
	case StepLobby:
		return "lobby"
	case StepVoting:
		return "voting"
	case StepResults:
		return "results"
	case StepFinished:
		return "finished"
	default:
		return "unknown"
	}
}

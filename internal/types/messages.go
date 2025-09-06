package types

// RoomMessage represents a structured message for room events (join/leave)
type RoomMessage struct {
	Subject  string `json:"subject"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// ChatMessage represents a chat message between users
type ChatMessage struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

// NatsPublishRequest represents a request to publish to NATS
type NatsPublishRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// MovieRequest represents a request containing movie IDs
type MovieRequest struct {
	MoviesReq []string `json:"moviesReq"`
}


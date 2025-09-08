package types

// Message represents a request to publish to NATS
type Message struct {
	Subject  string `json:"subject"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Room     string `json:"room"`
}

// MovieRequest represents a request containing movie IDs
type MovieRequest struct {
	MoviesReq []string `json:"moviesReq"`
}

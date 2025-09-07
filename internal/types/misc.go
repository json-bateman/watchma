package types

// Dumping types in this file until there's enough to refactor

// MovieReq represents an array of movie IDs
type MovieReq struct {
	MoviesReq []string `json:"movies"`
}

// Username represents a User and the room they're in
type Username struct {
	Username string `json:"username"`
	Roomname string `json:"roomname"`
}

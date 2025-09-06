package types

// Dumping types in this file until there's enough to refactor
type MovieReq struct {
	MoviesReq []string `json:"movies"`
}

type Username struct {
	Username string `json:"username"`
	Roomname string `json:"roomname"`
}

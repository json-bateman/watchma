package movie

// Movie is our internal representation of a movie that contains metadata expected on
// any movie, whether it comes from Jellyfin or Plex, etc.
type Movie struct {
	CommunityRating float64
	CriticRating    int
	Genres          []string
	Id              string
	Name            string
	OfficialRating  string
	PremiereDate    string
	PrimaryImageTag string
	ProductionYear  int
}

type SortField string

const (
	SortByName            SortField = "name"
	SortByYear            SortField = "year"
	SortByCriticRating    SortField = "critic"
	SortByCommunityRating SortField = "community"
)

type Query struct {
	SortBy     SortField
	Descending bool
	Genre      string
	Search     string
}

// MovieRequest represents a request containing movie IDs
type Request struct {
	Movies []string `json:"movies"`
}

// MovieVote represents a struct for holding final votes
type Vote struct {
	Movie *Movie
	Votes int
}

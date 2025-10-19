package types

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

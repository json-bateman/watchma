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

// CopySlice creates a deep copy of a movie slice to avoid shared references
func CopySlice(movies []Movie) []Movie {
	if movies == nil {
		return nil
	}

	copied := make([]Movie, len(movies))
	for i, m := range movies {
		// Copy the genres slice to avoid shared references
		genresCopy := make([]string, len(m.Genres))
		copy(genresCopy, m.Genres)

		copied[i] = Movie{
			CommunityRating: m.CommunityRating,
			CriticRating:    m.CriticRating,
			Genres:          genresCopy,
			Id:              m.Id,
			Name:            m.Name,
			OfficialRating:  m.OfficialRating,
			PremiereDate:    m.PremiereDate,
			PrimaryImageTag: m.PrimaryImageTag,
			ProductionYear:  m.ProductionYear,
		}
	}
	return copied
}

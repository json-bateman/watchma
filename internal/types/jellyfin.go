package types

type JellyfinItem struct {
	Name            string  `json:"Name"`
	Id              string  `json:"Id"`
	Container       string  `json:"Container"`
	PremiereDate    string  `json:"PremiereDate"`
	CriticRating    int     `json:"CriticRating"`
	CommunityRating float64 `json:"CommunityRating"`
	RunTimeTicks    int64   `json:"RunTimeTicks"`
	ProductionYear  int     `json:"ProductionYear"`
	ImageTags       struct {
		Primary string `json:"Primary"`
		Logo    string `json:"Logo"`
		Thumb   string `json:"Thumb"`
	} `json:"ImageTags"`
	BackdropImageTags []string `json:"BackdropImageTags"`
}

type JellyfinItems struct {
	Items []JellyfinItem `json:"Items"`
}

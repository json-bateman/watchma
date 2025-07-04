package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/json-bateman/jellyfin-grabber/internal"
)

type movieReq struct {
	MoviesReq []string `json:"movies"`
}

func PostMovies(w http.ResponseWriter, r *http.Request) {
	var movies movieReq
	fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&movies); err != nil {
		internal.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}
	if len(movies.MoviesReq) == 0 {
		internal.WriteJSONError(w, http.StatusBadRequest, "Must include at least 1 movie id.")
		return
	}
	fmt.Println(movies.MoviesReq)
}

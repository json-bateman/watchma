package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/json-bateman/jellyfin-grabber/internal"
	"github.com/json-bateman/jellyfin-grabber/view/movies"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type movieReq struct {
	MoviesReq []string `json:"movies"`
}

func PostMovies(w http.ResponseWriter, r *http.Request) {
	var moviesReq movieReq
	fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&moviesReq); err != nil {
		internal.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}
	if len(moviesReq.MoviesReq) == 0 {
		internal.WriteJSONError(w, http.StatusBadRequest, "Must include at least 1 movie id.")
		return
	}
	// TODO: Something with the movies

	// TODO: Maybe delete this and just redirect user to the room
	sse := datastar.NewSSE(w, r)
	sse.MergeFragmentTempl(movies.SubmitButton())

	fmt.Println(moviesReq.MoviesReq)
}

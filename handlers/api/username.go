package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
)

type Username struct {
	Username string `json:"username"`
	Roomname string `json:"roomname"`
}

var userSessions = make(map[string]string)

func SetUsername(w http.ResponseWriter, r *http.Request) {
	var u Username
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		SendSSEError(w, r, "Bad request", http.StatusBadRequest)
		return
	}

	if u.Username == "" {
		SendSSEError(w, r, "Please Enter a Username", http.StatusBadRequest)
		return
	}

	userSessions[u.Username] = u.Roomname

	fmt.Printf("User %s joined room %s\n", u.Username, u.Roomname)

	ClearSSEError(w, r)
	sse := datastar.NewSSE(w, r)
	if err := sse.PatchElementf(
		`<div id="inputs"> <p>Welcome, %s! You've joined room %s.</p> </div>`, u.Username, u.Roomname,
	); err != nil {
		fmt.Println("there was an error afoot", err)
		return
	}
}

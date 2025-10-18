package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"watchma/pkg/database/repository"

	"github.com/starfederation/datastar-go/datastar"
)

const userContextKey string = "user"

func WriteJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func WriteJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func GetSessionToken(r *http.Request) string {
	cookie, err := r.Cookie(SESSION_COOKIE_NAME)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func SendSSEError(w http.ResponseWriter, r *http.Request, message string) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElementf(`<div id="error" class="error-message text-lg text-red-500">%s</div>`, message)
}

func ClearSSEError(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElements(`<div id="error" class="hidden"></div>`)
}

// SetUserContext stores user data in the request context
func SetUserContext(r *http.Request, user *repository.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// GetUserFromContext retrieves user data from the request context
// Returns nil if no user is found in context
func GetUserFromContext(r *http.Request) *repository.User {
	user, ok := r.Context().Value(userContextKey).(*repository.User)
	if !ok {
		return nil
	}
	return user
}

func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	// Convert first rune (character) to upper case
	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

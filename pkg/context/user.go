package context

import (
	"context"
	"net/http"

	"watchma/db/repository"
)

type contextKey string

const userKey contextKey = "user"

// GetUser retrieves a user from the context
// Returns nil if no user is found
func GetUserFromRequest(r *http.Request) *repository.User {
	user, _ := r.Context().Value(userKey).(*repository.User)
	return user
}

// SetUserInRequest stores user in request context and returns the updated request
func SetUserInRequest(r *http.Request, user *repository.User) *http.Request {
	ctx := context.WithValue(r.Context(), userKey, user)
	return r.WithContext(ctx)
}

package context

import (
	"context"
	"net/http"

	"watchma/db/sqlcgen"
)

type contextKey string

const userKey contextKey = "user"

// GetUser retrieves a user from the context
// Returns nil if no user is found
func GetUserFromRequest(r *http.Request) *sqlcgen.User {
	user, _ := r.Context().Value(userKey).(*sqlcgen.User)
	return user
}

// SetUserInRequest stores user in request context and returns the updated request
func SetUserInRequest(r *http.Request, user *sqlcgen.User) *http.Request {
	ctx := context.WithValue(r.Context(), userKey, user)
	return r.WithContext(ctx)
}

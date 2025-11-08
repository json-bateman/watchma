package auth

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"watchma/pkg/auth"
	appctx "watchma/pkg/context"
)

// RequireLogin middleware checks for session cookie, loads user data, and stores in context
func RequireLogin(authService *auth.AuthService, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip middleware for static assets
			if strings.HasPrefix(r.URL.Path, "/public/") {
				next.ServeHTTP(w, r)
				return
			}

			token := getSessionToken(r)
			if token == "" {
				logger.Debug("User redirected to login, no session cookie")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			user, err := authService.GetUserBySessionToken(token)
			if err == sql.ErrNoRows {
				logger.Info("Session token exists but does not belong to a user. Session token invalid", "error", err.Error())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if err != nil {
				logger.Error("Failed to get user from session", "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Store user in request context
			r = appctx.SetUserInRequest(r, user)

			next.ServeHTTP(w, r)
		})
	}
}

// getSessionToken retrieves the session token from the request cookie
func getSessionToken(r *http.Request) string {
	cookie, err := r.Cookie(auth.SESSION_COOKIE_NAME)
	if err != nil {
		return ""
	}
	return cookie.Value
}

package web

import (
	"database/sql"
	"net/http"
	"strings"

	"watchma/pkg/utils"
)

// RequireLogin middleware checks for session cookie, loads user data, and stores in context
func (h *WebHandler) RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip middleware for static assets
		if strings.HasPrefix(r.URL.Path, "/public/") {
			next.ServeHTTP(w, r)
			return
		}

		// Get session token from cookie
		token := utils.GetSessionToken(r)
		if token == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Fetch user by session token (optimized JOIN query)
		user, err := h.authService.GetUserBySessionToken(token)
		if err == sql.ErrNoRows {
			// Invalid or expired session
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if err != nil {
			h.logger.Error("Failed to get user from session", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Store user in request context
		r = utils.SetUserContext(r, user)

		next.ServeHTTP(w, r)
	})
}

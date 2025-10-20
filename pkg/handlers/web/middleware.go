package web

import (
	"database/sql"
	"fmt"
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

		token := utils.GetSessionToken(r)
		if token == "" {
			h.logger.Debug(fmt.Sprintf("User redirected to login, no %s cookie", utils.SESSION_COOKIE_NAME))
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := h.services.AuthService.GetUserBySessionToken(token)
		if err == sql.ErrNoRows {
			h.logger.Info("User redirected to login, session token invalid", "error", err.Error())
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

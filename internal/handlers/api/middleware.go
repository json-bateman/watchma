package api

import (
	"net/http"
	"strings"

	"github.com/json-bateman/jellyfin-grabber/internal/utils"
)

// RequireUsername middleware checks for jelly_user cookie
func RequireUsername(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip middleware for static assets
		if strings.HasPrefix(r.URL.Path, "/public/") {
			next.ServeHTTP(w, r)
			return
		}

		username := utils.GetUsernameFromCookie(r)
		if username == "" {
			http.Redirect(w, r, "/username", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}


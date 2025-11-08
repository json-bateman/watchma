package web

import (
	"context"
	"encoding/json"
	"net/http"
	"watchma/pkg/database/repository"
	"watchma/pkg/types"

	"github.com/starfederation/datastar-go/datastar"
)

const UserContextKey string = "user"

func (h *WebHandler) SendSSEError(w http.ResponseWriter, r *http.Request, message string) {
	sse := datastar.NewSSE(w, r)
	if err := sse.PatchElementf(`<div id="error" class="error-message text-lg text-red-500">%s</div>`, message); err != nil {
		h.logger.Error("Failed to patch SSE error", "error", err, "message", message)
	}
}

func (h *WebHandler) ClearSSEError(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	if err := sse.PatchElements(`<div id="error" class="hidden"></div>`); err != nil {
		h.logger.Error("Failed to clear SSE error", "error", err)
	}
}

func (h *WebHandler) WriteJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (h *WebHandler) WriteJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *WebHandler) GetSessionToken(r *http.Request) string {
	cookie, err := r.Cookie(types.SESSION_COOKIE_NAME)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// SetUserContext stores user data in the request context
func (h *WebHandler) SetUserContext(r *http.Request, user *repository.User) *http.Request {
	ctx := context.WithValue(r.Context(), UserContextKey, user)
	return r.WithContext(ctx)
}

// GetUserFromContext retrieves user data from the request context
// Returns nil if no user is found in context
func (h *WebHandler) GetUserFromContext(r *http.Request) *repository.User {
	user, ok := r.Context().Value(UserContextKey).(*repository.User)
	if !ok {
		return nil
	}
	return user
}

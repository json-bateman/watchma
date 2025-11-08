package web

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
)

// WriteJSONError writes a JSON error response
func WriteJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// WriteJSONResponse writes a JSON response with the given data
func WriteJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// SendSSEError sends an error message via Server-Sent Events
func SendSSEError(w http.ResponseWriter, r *http.Request, message string, logger *slog.Logger) {
	sse := datastar.NewSSE(w, r)
	if err := sse.PatchElementf(`<div id="error" class="error-message text-lg text-red-500">%s</div>`, message); err != nil {
		if logger != nil {
			logger.Error("Failed to patch SSE error", "error", err, "message", message)
		}
	}
}

// ClearSSEError clears the error message element via Server-Sent Events
func ClearSSEError(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	sse := datastar.NewSSE(w, r)
	if err := sse.PatchElements(`<div id="error" class="hidden"></div>`); err != nil {
		if logger != nil {
			logger.Error("Failed to clear SSE error", "error", err)
		}
	}
}

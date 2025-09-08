package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
)

func PrintRes(resp *http.Response) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("error reading response body: %v", err.Error()))
		return
	}
	bodyString := string(bodyBytes)
	fmt.Println("HTML response body:")
	fmt.Println(bodyString)
}

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

func GetUsernameFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(USERNAME_COOKIE)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func SendSSEError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElementf(`<div id="error" class="error-message text-red-500">%s</div>`, message)
}

func ClearSSEError(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElements(`<div id="error" class="hidden"></div>`)
}

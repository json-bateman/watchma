package api

import (
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
)

func SendSSEError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElementf(`<div id="error" class="error-message text-red-500">%s</div>`, message)
}

func ClearSSEError(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElements(`<div id="error" class="hidden"></div>`)
}
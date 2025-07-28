package pages

import (
	"net/http"
	"time"

	"github.com/starfederation/datastar-go/datastar"
)

func TestSSE(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	for {
		select {
		case <-r.Context().Done():
			return // <- THIS closes the response!
		default:
			sse.Send("ping", []string{`<div class="text-blue-300">yolo</div>`})
			time.Sleep(200 * time.Millisecond)
		}
	}
}

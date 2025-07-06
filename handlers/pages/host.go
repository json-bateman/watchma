package pages

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/json-bateman/jellyfin-grabber/view/host"
)

func Host(w http.ResponseWriter, r *http.Request) {
	component := host.HostPage()
	templ.Handler(component).ServeHTTP(w, r)
}

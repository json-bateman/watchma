package components

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/json-bateman/jellyfin-grabber/view"
)

func Index(w http.ResponseWriter, r *http.Request) {
	component := view.Index("Sup wit it")
	templ.Handler(component).ServeHTTP(w, r)
}

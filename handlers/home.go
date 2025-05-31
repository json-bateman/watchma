package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/json-bateman/jellyfin-grabber/view/home"
)

type HomeHandler struct{}

func (h HomeHandler) Show(w http.ResponseWriter, r *http.Request) {
	component := home.Home()
	templ.Handler(component).ServeHTTP(w, r)
}

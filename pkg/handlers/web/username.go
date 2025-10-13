package web

import (
	"net/http"

	"github.com/a-h/templ"
	"watchma/pkg/utils"
	"watchma/view/username"
)

func (h *WebHandler) Username(w http.ResponseWriter, r *http.Request) {
	component := username.UsernameForm()
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) SetUsername(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")

	http.SetCookie(w, &http.Cookie{
		Name:   utils.USERNAME_COOKIE,
		Value:  username,
		Path:   "/",
		MaxAge: 30 * 24 * 60 * 60, // 30 days
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

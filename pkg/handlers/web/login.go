package web

import (
	"net/http"

	"watchma/view/login"

	"github.com/a-h/templ"
)

func (h *WebHandler) Login(w http.ResponseWriter, r *http.Request) {
	component := login.LoginForm()
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *WebHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validate input
	if username == "" || password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	// Login or create user
	user, token, err := h.authService.LoginOrCreate(username, password)
	if err != nil {
		h.logger.Error("Login failed", "error", err, "username", username)
		http.Error(w, "Login failed", http.StatusUnauthorized)
		return
	}

	// Set session token as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	h.logger.Info("Login successful", "user_id", user.ID, "username", user.Username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

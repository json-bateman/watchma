package web

import (
	"encoding/json"
	"net/http"
	"unicode"

	"watchma/pkg/utils"
	"watchma/view/common"
	"watchma/view/login"

	"github.com/starfederation/datastar-go/datastar"
)

type Signals struct {
	Password string `json:"password"`
}

func valid(pw string) (login.PwRules, bool) {
	var rules login.PwRules

	runes := []rune(pw)
	n := len(runes)

	rules.Has8 = n >= 8

	for _, r := range runes {
		switch {
		case unicode.IsLower(r):
			rules.HasLower = true
		case unicode.IsUpper(r):
			rules.HasUpper = true
		case unicode.IsDigit(r):
			rules.HasNumber = true
		}
	}

	valid := rules.Has8 && rules.HasLower && rules.HasUpper && rules.HasNumber
	return rules, valid

}

func (h *WebHandler) ValidatePassword(w http.ResponseWriter, r *http.Request) {
	var signals Signals
	if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
		h.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}

	rules, _ := valid(signals.Password)

	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(login.LoginForm(rules))
}

func (h *WebHandler) Login(w http.ResponseWriter, r *http.Request) {
	component := login.LoginForm(login.PwRules{
		Has8:      false,
		HasLower:  false,
		HasUpper:  false,
		HasNumber: false,
	})
	h.RenderPageNoLayout(component, "Movie Showdown", w, r)
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

	_, valid := valid(password)

	if !valid {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(common.Error("Password is Invalid!"))
		return
	}

	user, token, err := h.services.AuthService.LoginOrCreate(username, password)
	if err != nil {
		sse := datastar.NewSSE(w, r)
		h.logger.Error("Login failed", "error", err, "username", username)
		sse.PatchElementTempl(common.Error(err.Error()))
		return
	}

	// Set session token as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     utils.SESSION_COOKIE_NAME,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	sse := datastar.NewSSE(w, r)
	h.logger.Info("Login successful", "user_id", user.ID, "username", user.Username)
	sse.Redirect("/")
}

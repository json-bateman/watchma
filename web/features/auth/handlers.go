package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"unicode"
	"watchma/pkg/auth"
	"watchma/web"
	"watchma/web/features/auth/pages"
	"watchma/web/views/common"

	"github.com/starfederation/datastar-go/datastar"
)

type handlers struct {
	authService *auth.AuthService
	logger      *slog.Logger
}

func newHandlers(auth *auth.AuthService, l *slog.Logger) *handlers {
	return &handlers{
		logger:      l,
		authService: auth,
	}
}

type Signals struct {
	Password string `json:"password"`
}

func valid(pw string) (pages.PwRules, bool) {
	var rules pages.PwRules

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

func (h *handlers) ValidatePassword(w http.ResponseWriter, r *http.Request) {
	var signals Signals
	if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
		web.WriteJSONError(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}

	rules, _ := valid(signals.Password)

	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(pages.LoginForm(rules))
}

func (h *handlers) Login(w http.ResponseWriter, r *http.Request) {
	component := pages.LoginForm(pages.PwRules{
		Has8:      false,
		HasLower:  false,
		HasUpper:  false,
		HasNumber: false,
	})
	web.RenderPageNoLayout(component, "Watchma", w, r)
}

func (h *handlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
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

	user, token, err := h.authService.LoginOrCreate(username, password)
	if err != nil {
		sse := datastar.NewSSE(w, r)
		h.logger.Warn("Login failed", "error", err, "username", username)
		sse.PatchElementTempl(common.Error(err.Error()))
		return
	}

	// Set session token as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   !h.authService.IsDev,
		SameSite: http.SameSiteLaxMode,
	})

	sse := datastar.NewSSE(w, r)
	h.logger.Debug("Login successful", "user_id", user.ID, "username", user.Username)
	sse.Redirect("/")
}

func (h *handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   !h.authService.IsDev,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // This deletes the cookie
	})

	// Redirect to login page
	sse := datastar.NewSSE(w, r)
	h.logger.Debug("Logout successful")
	sse.Redirect("/")
}

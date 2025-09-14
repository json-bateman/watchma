package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUsernameFromCookie_MissingCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	username := GetUsernameFromCookie(req)
	assert.Equal(t, "", username)
}

func TestGetUsernameFromCookie_WrongCookieName(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	cookie := &http.Cookie{
		Name:  "that_is_not_my_cookie_fr",
		Value: "testuser123",
	}
	req.AddCookie(cookie)

	username := GetUsernameFromCookie(req)
	assert.Equal(t, "", username)
}

func TestGetUsernameFromCookie_EmptyValue(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	cookie := &http.Cookie{
		Name:  USERNAME_COOKIE,
		Value: "",
	}
	req.AddCookie(cookie)

	username := GetUsernameFromCookie(req)
	assert.Equal(t, "", username)
}

func TestGetUsernameFromCookie_MultipleCookies(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	cookie1 := &http.Cookie{
		Name:  "random_ad_tracking_cookie",
		Value: "probably_from_meta",
	}
	cookie2 := &http.Cookie{
		Name:  USERNAME_COOKIE,
		Value: "the_real_jelly_user_stand_up",
	}
	cookie3 := &http.Cookie{
		Name:  "session",
		Value: "session_value",
	}

	req.AddCookie(cookie1)
	req.AddCookie(cookie2)
	req.AddCookie(cookie3)

	username := GetUsernameFromCookie(req)
	assert.Equal(t, "the_real_jelly_user_stand_up", username)
}

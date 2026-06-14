package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewToken_Length(t *testing.T) {
	token := NewToken()
	// 32 bytes = 64 hex chars
	if len(token) != 64 {
		t.Errorf("expected token length 64, got %d", len(token))
	}
}

func TestNewToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		tk := NewToken()
		if tokens[tk] {
			t.Fatal("duplicate token generated")
		}
		tokens[tk] = true
	}
}

func TestSetSessionCookie(t *testing.T) {
	w := httptest.NewRecorder()
	SetSessionCookie(w, "test-token-abc")

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected cookie to be set")
	}

	c := cookies[0]
	if c.Name != sessionCookie {
		t.Errorf("expected cookie name %q, got %q", sessionCookie, c.Name)
	}
	if c.Value != "test-token-abc" {
		t.Errorf("expected cookie value %q, got %q", "test-token-abc", c.Value)
	}
	if !c.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Error("expected SameSite Lax")
	}
	if c.MaxAge != int(sessionDuration.Seconds()) {
		t.Errorf("expected MaxAge %d, got %d", int(sessionDuration.Seconds()), c.MaxAge)
	}
}

func TestClearSessionCookie(t *testing.T) {
	w := httptest.NewRecorder()
	ClearSessionCookie(w)

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected cookie to be set")
	}

	c := cookies[0]
	if c.Name != sessionCookie {
		t.Errorf("expected cookie name %q, got %q", sessionCookie, c.Name)
	}
	if c.MaxAge != -1 {
		t.Errorf("expected MaxAge -1, got %d", c.MaxAge)
	}
}

func TestGetUserFromSession_NoCookie(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	// nil db is fine — should return nil before hitting DB
	u := GetUserFromSession(nil, r)
	if u != nil {
		t.Error("expected nil user when no cookie")
	}
}

func TestProtected_RedirectsWithoutSession(t *testing.T) {
	called := false
	handler := Protected(nil, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	r := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	handler(w, r)

	if called {
		t.Error("next handler should not have been called")
	}
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected redirect 303, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "/login" {
		t.Errorf("expected redirect to /login, got %q", loc)
	}
}

func TestAdminOnly_RedirectsWithoutSession(t *testing.T) {
	called := false
	handler := AdminOnly(nil, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	r := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	handler(w, r)

	if called {
		t.Error("next handler should not have been called")
	}
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected redirect 303, got %d", w.Code)
	}
}

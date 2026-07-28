package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAuthRegisterLogin(t *testing.T) {
	srv, mux := setupTestServer(t)

	// Test Registration
	form := url.Values{}
	form.Add("username", "testuser")
	form.Add("password", "testpass")

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected 200 OK with HX-Redirect after register, got %v", status)
	}

	// Verify user exists
	u, err := srv.DB.AuthenticateUser("testuser", "testpass")
	if err != nil {
		t.Fatalf("failed to authenticate registered user: %v", err)
	}
	if u.Username != "testuser" {
		t.Errorf("expected testuser, got %v", u.Username)
	}

	// Test Login
	reqLogin := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(form.Encode()))
	reqLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rrLogin := httptest.NewRecorder()

	mux.ServeHTTP(rrLogin, reqLogin)

	if status := rrLogin.Code; status != http.StatusOK {
		t.Errorf("expected 200 OK with HX-Redirect after login, got %v", status)
	}
}

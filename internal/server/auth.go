package server

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderError(w, r, "Bad Request", http.StatusBadRequest)
		return
	}
	user := s.getUserFromSession(r)
	if user == nil {
		s.renderError(w, r, "Authentication required.", http.StatusUnauthorized)
		return
	}

	handle := strings.TrimSpace(r.FormValue("handle"))
	if handle == "" {
		handle = strings.TrimSpace(r.FormValue("username"))
	}
	if handle == "" {
		s.renderError(w, r, "Handle required", http.StatusBadRequest)
		return
	}

	// Format validation for handle
	matched, err := regexp.MatchString(`^[^\s#]+#\d{4}$`, handle)
	if err != nil || !matched {
		hashIndex := strings.Index(handle, "#")
		baseName := handle
		if hashIndex != -1 {
			baseName = strings.TrimSpace(handle[:hashIndex])
		}
		if baseName == "" || strings.Contains(baseName, " ") {
			s.renderError(w, r, "Invalid Operator handle format. No spaces allowed.", http.StatusBadRequest)
			return
		}
		num := (time.Now().UnixNano() % 9000) + 1000
		handle = fmt.Sprintf("%s#%04d", baseName, num)
	}

	user.Handle = handle
	if err := s.DB.SaveUser(user); err != nil {
		s.renderError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	setCookie(w, "cy_username", user.Handle)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderError(w, r, "Bad Request", http.StatusBadRequest)
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := strings.TrimSpace(r.FormValue("password"))

	if username == "" || password == "" {
		s.redirect(w, r, "/?error="+url.QueryEscape("Login username and password required."))
		return
	}

	if strings.Contains(username, " ") {
		s.redirect(w, r, "/?error="+url.QueryEscape("Username cannot contain spaces."))
		return
	}

	u, err := s.DB.CreateUserAccount(username, password)
	if err != nil {
		s.redirect(w, r, "/?error="+url.QueryEscape(err.Error()))
		return
	}

	setCookie(w, "cy_user_id", u.ID)
	setCookie(w, "cy_username", u.Handle)

	s.redirect(w, r, "/")
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderError(w, r, "Bad Request", http.StatusBadRequest)
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := strings.TrimSpace(r.FormValue("password"))

	if username == "" || password == "" {
		s.redirect(w, r, "/?error="+url.QueryEscape("Username and password required."))
		return
	}

	u, err := s.DB.AuthenticateUser(username, password)
	if err != nil {
		s.redirect(w, r, "/?error="+url.QueryEscape(err.Error()))
		return
	}

	setCookie(w, "cy_user_id", u.ID)
	setCookie(w, "cy_username", u.Handle)

	s.redirect(w, r, "/")
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "cy_user_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "cy_username",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	s.redirect(w, r, "/")
}

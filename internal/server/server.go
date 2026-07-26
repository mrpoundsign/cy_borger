package server

import (
	"html/template"
	"net/http"

	"github.com/mrpoundsign/cy_borger/internal/db"
)

type Server struct {
	DB        *db.DB
	Templates *template.Template
}

func NewServer(database *db.DB, templates *template.Template) *Server {
	return &Server{
		DB:        database,
		Templates: templates,
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("POST /auth/register", s.handleRegister)
	mux.HandleFunc("POST /auth/login", s.handleLogin)
	mux.HandleFunc("POST /auth/logout", s.handleLogout)
	mux.HandleFunc("POST /user/update", s.handleUpdateUser)
	
	mux.HandleFunc("POST /character/generate", s.handleGenerateCharacter)
	mux.HandleFunc("POST /character/create_blank", s.handleCreateBlankCharacter)
	mux.HandleFunc("GET /character/{id}", s.handleViewCharacter)
	mux.HandleFunc("POST /character/{id}/auth", s.handleAuthCharacter)
	mux.HandleFunc("POST /character/{id}/join", s.handleJoinGame)
	mux.HandleFunc("POST /character/{id}/keep", s.handleKeepCharacter)
	mux.HandleFunc("PUT /character/{id}/update_hp", s.handleUpdateHP)
	mux.HandleFunc("PUT /character/{id}/update_glitches", s.handleUpdateGlitches)
	mux.HandleFunc("PUT /character/{id}/update_stat", s.handleUpdateStat)
	mux.HandleFunc("PUT /character/{id}/update_field", s.handleUpdateField)
	mux.HandleFunc("POST /character/{id}/update_field", s.handleUpdateField)
	mux.HandleFunc("POST /character/{id}/add_item", s.handleAddListItem)
	mux.HandleFunc("POST /character/{id}/delete_item", s.handleDeleteListItem)
	mux.HandleFunc("POST /character/{id}/delete", s.handleDeleteCharacter)
	mux.HandleFunc("POST /character/{id}/kill", s.handleKillCharacter)
	mux.HandleFunc("POST /character/{id}/revive", s.handleReviveCharacter)

	mux.HandleFunc("POST /game/create", s.handleCreateGame)
	mux.HandleFunc("GET /game/{id}", s.handleViewGame)
	mux.HandleFunc("POST /game/{id}/auth", s.handleAuthGame)

	// WebSockets
	mux.HandleFunc("GET /ws/game/{id}", s.handleWSGame)
	mux.HandleFunc("GET /ws/character/{id}", s.handleWSCharacter)
}

// Helpers for Session Cookies
func setCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func getCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (s *Server) getUserFromSession(r *http.Request) *db.User {
	cookie, err := r.Cookie("cy_user_id")
	if err != nil || cookie.Value == "" {
		return nil
	}
	u, err := s.DB.GetUser(cookie.Value)
	if err != nil || u == nil {
		return nil
	}
	return u
}

func (s *Server) renderError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	if r.Header.Get("HX-Request") == "true" {
		http.Error(w, message, statusCode)
		return
	}
	w.WriteHeader(statusCode)
	data := map[string]interface{}{
		"ErrorMessage": message,
		"StatusCode":   statusCode,
	}
	_ = s.Templates.ExecuteTemplate(w, "error.html", data)
}

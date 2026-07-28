package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mrpoundsign/cy_borger/internal/db"
	"github.com/mrpoundsign/cy_borger/internal/templates"
)

type Server struct {
	DB *db.DB
}

func NewServer(database *db.DB) *Server {
	return &Server{
		DB: database,
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
	mux.HandleFunc("GET /game/{id}/party", s.handleGameParty)
	mux.HandleFunc("GET /game/{id}/logs", s.handleGetGameLogs)
	mux.HandleFunc("POST /game/{id}/auth", s.handleAuthGame)

	// WebSockets
	mux.HandleFunc("GET /ws/game/{id}", s.handleWSGame)

	// GM Controls
	mux.HandleFunc("POST /game/{id}/rename", s.handleRenameGame)
	mux.HandleFunc("POST /game/{id}/toggle_lock", s.handleToggleLock)
	mux.HandleFunc("POST /game/{id}/kick", s.handleKickUser)
	mux.HandleFunc("POST /game/{id}/ban", s.handleBanUser)
	mux.HandleFunc("POST /game/{id}/promote_gm", s.handlePromoteGM)
	mux.HandleFunc("POST /game/{id}/demote_gm", s.handleDemoteGM)
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

func (s *Server) redirect(w http.ResponseWriter, r *http.Request, target string) {
	w.Header().Set("HX-Redirect", target)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) renderError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	if r.Header.Get("HX-Request") == "true" {
		http.Error(w, message, statusCode)
		return
	}
	w.WriteHeader(statusCode)
	if err := templates.Base(fmt.Sprintf("CY_BORGER - Error %d", statusCode), nil, templates.Error(statusCode, message)).Render(r.Context(), w); err != nil {
		log.Printf("Template execution error (error.templ): %v", err)
	}
}

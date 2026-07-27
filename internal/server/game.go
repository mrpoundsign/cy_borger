package server

import (
	"log"
	"net/http"

	"github.com/mrpoundsign/cy_borger/templates"
)

func (s *Server) handleCreateGame(w http.ResponseWriter, r *http.Request) {
	user := s.getUserFromSession(r)
	if user == nil {
		s.renderError(w, r, "Authentication required. Please log in or register an account.", http.StatusUnauthorized)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		name = "Unnamed Campaign"
	}

	ownerID := user.ID

	g, err := s.DB.CreateGame(name, ownerID)
	if err != nil {
		s.renderError(w, r, "Failed to create game", http.StatusInternalServerError)
		return
	}

	setCookie(w, "game_gm_"+g.ID, g.GMCode)
	setCookie(w, "last_game_invite", g.InviteCode)

	w.Header().Set("HX-Redirect", "/game/"+g.ID)
	http.Redirect(w, r, "/game/"+g.ID, http.StatusSeeOther)
}

func (s *Server) handleViewGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if gmCode := r.URL.Query().Get("gm_code"); gmCode != "" {
		setCookie(w, "game_gm_"+id, gmCode)
		http.Redirect(w, r, "/game/"+id, http.StatusSeeOther)
		return
	}

	g, err := s.DB.GetGame(id)
	if err != nil || g == nil {
		http.NotFound(w, r)
		return
	}

	chars, err := s.DB.GetCharactersForGame(g.ID)
	if err != nil {
		log.Printf("Failed to get characters for game %s: %v", g.ID, err)
	}

	// Compute latest update timestamp among game and its characters
	latestUpdate := g.UpdatedAt
	for _, c := range chars {
		if c.UpdatedAt.After(latestUpdate) {
			latestUpdate = c.UpdatedAt
		}
	}

	// HTTP Caching Headers
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	w.Header().Set("Vary", "HX-Request, Accept")
	w.Header().Set("Last-Modified", latestUpdate.UTC().Format(http.TimeFormat))

	user := s.getUserFromSession(r)
	isGM := (getCookie(r, "game_gm_"+g.ID) == g.GMCode) || (user != nil && user.ID != "" && g.OwnerID == user.ID)

	currentUserID := getCookie(r, "cy_user_id")
	if currentUserID == "" && user != nil {
		currentUserID = user.ID
	}
	logs, err := s.DB.GetGameLogs(g.ID, nil, 50)
	if err != nil {
		log.Printf("Failed to get logs for game %s: %v", g.ID, err)
	}

	if err := templates.Base("CY_BORGER - GAME", nil, templates.Game(g, chars, isGM, currentUserID, logs)).Render(r.Context(), w); err != nil {
		log.Printf("Template execution error (game.templ): %v", err)
	}
}

func (s *Server) handleGameParty(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	game, err := s.DB.GetGame(id)
	if err != nil || game == nil {
		http.NotFound(w, r)
		return
	}
	characters, err := s.DB.GetCharactersForGame(game.ID)
	if err != nil {
		log.Printf("Failed to get characters for game %s: %v", game.ID, err)
	}
	user := s.getUserFromSession(r)
	isGM := (getCookie(r, "game_gm_"+game.ID) == game.GMCode) || (user != nil && user.ID != "" && game.OwnerID == user.ID)
	currentUserID := getCookie(r, "cy_user_id")
	if currentUserID == "" && user != nil {
		currentUserID = user.ID
	}
	if err := templates.PartyGrid(game, characters, isGM, currentUserID).Render(r.Context(), w); err != nil {
		log.Printf("Template execution error (party_grid.templ): %v", err)
	}
}

func (s *Server) handleAuthGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	gmCode := r.FormValue("gm_code")

	g, err := s.DB.GetGame(id)
	if err == nil && g != nil && g.GMCode == gmCode {
		setCookie(w, "game_gm_"+id, gmCode)
		setCookie(w, "last_game_invite", g.InviteCode)
	}

	w.Header().Set("HX-Redirect", "/game/"+id)
	http.Redirect(w, r, "/game/"+id, http.StatusSeeOther)
}

func (s *Server) handleGetGameLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		s.renderError(w, r, "Invalid form data", http.StatusBadRequest)
		return
	}

	g, err := s.DB.GetGame(id)
	if err != nil || g == nil {
		http.NotFound(w, r)
		return
	}

	types := r.Form["type"]
	logs, err := s.DB.GetGameLogs(g.ID, types, 50)
	if err != nil {
		log.Printf("Failed to get logs for game %s: %v", g.ID, err)
	}

	chars, err := s.DB.GetCharactersForGame(g.ID)
	if err != nil {
		log.Printf("Failed to get characters for game %s: %v", g.ID, err)
	}

	if err := templates.GameLogs(g, chars, logs).Render(r.Context(), w); err != nil {
		log.Printf("Template execution error (game_logs.templ): %v", err)
	}
}

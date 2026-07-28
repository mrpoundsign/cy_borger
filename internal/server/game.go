package server

import (
	"log"
	"net/http"

	"github.com/mrpoundsign/cy_borger/internal/db"
	"github.com/mrpoundsign/cy_borger/internal/templates"
)

func (s *Server) handleCreateGame(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderError(w, r, "Bad Request", http.StatusBadRequest)
		return
	}
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
	w.WriteHeader(http.StatusOK)
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

	// Fetch GMs to see if current user is an additional GM explicitly
	gms, _ := s.DB.GetGameGMs(g.ID)
	if !isGM && user != nil {
		for _, gm := range gms {
			if gm.ID == user.ID {
				isGM = true
				break
			}
		}
	}

	currentUserID := getCookie(r, "cy_user_id")
	if currentUserID == "" && user != nil {
		currentUserID = user.ID
	}
	logs, err := s.DB.GetGameLogs(g.ID, nil, 50)
	if err != nil {
		log.Printf("Failed to get logs for game %s: %v", g.ID, err)
	}

	// For the Operators UI, we need a unique list of players in this game
	playerMap := make(map[string]db.User)
	for _, c := range chars {
		if c.OwnerID != "" && c.OwnerUsername != "" {
			playerMap[c.OwnerID] = db.User{ID: c.OwnerID, Username: c.OwnerUsername}
		}
	}
	var players []db.User
	for _, p := range playerMap {
		players = append(players, p)
	}

	if err := templates.Base("CY_BORGER - GAME", nil, templates.Game(g, chars, isGM, currentUserID, logs, players, gms)).Render(r.Context(), w); err != nil {
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
	if err := r.ParseForm(); err != nil {
		s.renderError(w, r, "Bad Request", http.StatusBadRequest)
		return
	}
	id := r.PathValue("id")
	gmCode := r.FormValue("gm_code")

	g, err := s.DB.GetGame(id)
	if err == nil && g != nil && g.GMCode == gmCode {
		setCookie(w, "game_gm_"+id, gmCode)
		setCookie(w, "last_game_invite", g.InviteCode)

		user := s.getUserFromSession(r)
		if user != nil && user.ID != "" {
			_ = s.DB.AddGameGM(g.ID, user.ID)
		}
	}

	w.Header().Set("HX-Redirect", "/game/"+id)
	w.WriteHeader(http.StatusOK)
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

func (s *Server) handleRenameGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	name := r.FormValue("name")

	user := s.getUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	g, err := s.DB.GetGame(id)
	if err != nil || g == nil {
		http.NotFound(w, r)
		return
	}

	if g.OwnerID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := s.DB.RenameGame(id, name); err != nil {
		http.Error(w, "Failed to rename", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleToggleLock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user := s.getUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	g, err := s.DB.GetGame(id)
	if err != nil || g == nil {
		http.NotFound(w, r)
		return
	}

	if g.OwnerID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	isLocked := r.FormValue("is_locked") == "true"
	if err := s.DB.ToggleGameLock(id, isLocked); err != nil {
		http.Error(w, "Failed to toggle lock", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleKickUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.FormValue("user_id")

	user := s.getUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	g, err := s.DB.GetGame(id)
	if err != nil || g == nil {
		http.NotFound(w, r)
		return
	}

	// Only owner or GM can kick
	isGM := g.OwnerID == user.ID
	if !isGM {
		gms, _ := s.DB.GetGameGMs(g.ID)
		for _, gm := range gms {
			if gm.ID == user.ID {
				isGM = true
				break
			}
		}
	}

	if !isGM {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := s.DB.KickUserFromGame(id, userID); err != nil {
		http.Error(w, "Failed to kick", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleBanUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	bannedUserID := r.FormValue("user_id")

	user := s.getUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	g, err := s.DB.GetGame(id)
	if err != nil || g == nil {
		http.NotFound(w, r)
		return
	}

	isGM := g.OwnerID == user.ID
	if !isGM {
		gms, _ := s.DB.GetGameGMs(g.ID)
		for _, gm := range gms {
			if gm.ID == user.ID {
				isGM = true
				break
			}
		}
	}

	if !isGM {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := s.DB.BanUser(g.OwnerID, bannedUserID); err != nil {
		http.Error(w, "Failed to ban", http.StatusInternalServerError)
		return
	}

	if err := s.DB.KickUserFromAllOwnerGames(g.OwnerID, bannedUserID); err != nil {
		http.Error(w, "Failed to kick from games", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePromoteGM(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.FormValue("user_id")

	user := s.getUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	g, err := s.DB.GetGame(id)
	if err != nil || g == nil {
		http.NotFound(w, r)
		return
	}

	if g.OwnerID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := s.DB.AddGameGM(id, userID); err != nil {
		http.Error(w, "Failed to promote", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDemoteGM(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.FormValue("user_id")

	user := s.getUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	g, err := s.DB.GetGame(id)
	if err != nil || g == nil {
		http.NotFound(w, r)
		return
	}

	if g.OwnerID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := s.DB.RemoveGameGM(id, userID); err != nil {
		http.Error(w, "Failed to demote", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeleteGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user := s.getUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	g, err := s.DB.GetGame(id)
	if err != nil || g == nil {
		http.NotFound(w, r)
		return
	}

	if g.OwnerID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := s.DB.DeleteGame(id); err != nil {
		http.Error(w, "Failed to delete game", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

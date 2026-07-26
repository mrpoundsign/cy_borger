package server

import (
	"log"
	"net/http"
	"time"
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

	chars, _ := s.DB.GetCharactersForGame(g.ID)

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

	// Check If-Modified-Since header (HTTP Caching)
	if r.Header.Get("HX-Request") != "true" {
		if ifModSince := r.Header.Get("If-Modified-Since"); ifModSince != "" {
			if t, err := time.Parse(http.TimeFormat, ifModSince); err == nil {
				if !latestUpdate.Truncate(time.Second).After(t) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}
	}

	user := s.getUserFromSession(r)
	isGM := (getCookie(r, "game_gm_"+g.ID) == g.GMCode) || (user != nil && user.ID != "" && g.OwnerID == user.ID)

	currentUserID := getCookie(r, "cy_user_id")
	if currentUserID == "" && user != nil {
		currentUserID = user.ID
	}

	data := map[string]interface{}{
		"Game":          g,
		"Characters":    chars,
		"IsGM":          isGM,
		"CurrentUserID": currentUserID,
	}

	if err := s.Templates.ExecuteTemplate(w, "game.html", data); err != nil {

		log.Printf("Template execution error (game.html): %v", err)

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

	http.Redirect(w, r, "/game/"+id, http.StatusSeeOther)
}

package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mrpoundsign/cy_borger/internal/db"
	"github.com/mrpoundsign/cy_borger/internal/ws"
	"github.com/mrpoundsign/cy_borger/pkg/chargen"
)

func (s *Server) logAndBroadcastGameEvent(c *chargen.Character, eventType, message string) {
	if err := s.DB.LogGameEvent(c.GameID, c.ID, eventType, message); err != nil {
		log.Printf("Failed to log event: %v", err)
	}
	if c.GameID != "" {
		charName := c.Name
		if charName == "" {
			charName = "[UNNAMED OPERATOR]"
		}
		msgB64 := base64.StdEncoding.EncodeToString([]byte(message))
		ws.GlobalHub.Broadcast("game_"+c.GameID, fmt.Sprintf("log_entry:%s:%s:%s", eventType, charName, msgB64))
	}
}

func (s *Server) handleGenerateCharacter(w http.ResponseWriter, r *http.Request) {
	user := s.getUserFromSession(r)
	if user == nil {
		s.renderError(w, r, "Authentication required. Please log in or register an account.", http.StatusUnauthorized)
		return
	}

	c := chargen.GenerateCharacter()
	ownerID := user.ID

	if user.Handle != "" {
		c.Handle = user.Handle
	}

	if err := r.ParseForm(); err != nil {

		log.Printf("Failed to parse form: %v", err)

		http.Error(w, "Bad Request", http.StatusBadRequest)

		return

	}
	if gameID := r.FormValue("game_id"); gameID != "" {
		c.GameID = gameID
	}

	if err := s.DB.SaveCharacter(&c, ownerID); err != nil {
		s.renderError(w, r, "Failed to save character", http.StatusInternalServerError)
		return
	}

	// Set edit session cookie automatically for creator
	setCookie(w, "char_edit_"+c.ID, c.EditCode)

	// If HTMX request, render the character template directly with HX-Push-Url header!
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Push-Url", "/character/"+c.ID)
		s.renderCharacterViewWithChar(w, r, &c)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func (s *Server) renderCharacterViewWithChar(w http.ResponseWriter, r *http.Request, c *chargen.Character) {
	user := s.getUserFromSession(r)
	sessionCode := getCookie(r, "char_edit_"+c.ID)
	canEdit := (sessionCode != "" && sessionCode == c.EditCode) || (user != nil && user.ID != "" && c.OwnerID == user.ID)

	var game *db.Game
	isGM := false
	if c.GameID != "" {
		g, err := s.DB.GetGame(c.GameID)
		if err != nil {
			log.Printf("Failed to get game %s: %v", c.GameID, err)
		} else {
			game = g
		}
		if game != nil {
			isGM = getCookie(r, "game_gm_"+game.ID) == game.GMCode
			if !canEdit && isGM {
				canEdit = true
			}
		}
	}

	// Check remembered active game invite if character not in game
	var activeGame *db.Game
	if game != nil {
		activeGame = game
	} else if lastInvite := getCookie(r, "last_game_invite"); lastInvite != "" {
		ag, err := s.DB.GetGameByInviteCode(lastInvite)
		if err != nil {
			log.Printf("Failed to get game by invite code %s: %v", lastInvite, err)
		} else {
			activeGame = ag
		}
	}

	data := map[string]interface{}{
		"Character":  c,
		"CanEdit":    canEdit,
		"IsGM":       isGM,
		"Game":       game,
		"ActiveGame": activeGame,
		"IsModal":    r.URL.Query().Get("modal") == "true",
	}

	// Render in modal if requested
	if r.URL.Query().Get("modal") == "true" {
		if err := s.Templates.ExecuteTemplate(w, "character_modal.html", data); err != nil {
			log.Printf("Template execution error (character_modal.html): %v", err)
		}
		return
	}

	if err := s.Templates.ExecuteTemplate(w, "character.html", data); err != nil {

		log.Printf("Template execution error (character.html): %v", err)

	}
}

func (s *Server) handleViewCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if code := r.URL.Query().Get("code"); code != "" {
		// Set cookie & redirect to clean URL
		setCookie(w, "char_edit_"+id, code)
		http.Redirect(w, r, "/character/"+id, http.StatusSeeOther)
		return
	}

	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		http.NotFound(w, r)
		return
	}

	// HTTP Caching Headers
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	w.Header().Set("Vary", "HX-Request, Accept")
	w.Header().Set("Last-Modified", c.UpdatedAt.UTC().Format(http.TimeFormat))

	// Check If-Modified-Since header (HTTP Caching)
	if r.Header.Get("HX-Request") != "true" && r.URL.Query().Get("modal") != "true" {
		if ifModSince := r.Header.Get("If-Modified-Since"); ifModSince != "" {
			if t, err := time.Parse(http.TimeFormat, ifModSince); err == nil {
				if !c.UpdatedAt.Truncate(time.Second).After(t) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}
	}

	s.renderCharacterViewWithChar(w, r, c)
}

func (s *Server) handleAuthCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	editCode := r.FormValue("edit_code")

	c, err := s.DB.GetCharacter(id)
	if err == nil && c != nil && c.EditCode == editCode {
		setCookie(w, "char_edit_"+id, editCode)
	}

	http.Redirect(w, r, "/character/"+id, http.StatusSeeOther)
}

func (s *Server) handleJoinGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inviteCode := r.FormValue("invite_code")

	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		http.NotFound(w, r)
		return
	}

	g, err := s.DB.GetGameByInviteCode(inviteCode)
	if err != nil || g == nil {
		http.Redirect(w, r, "/character/"+id+"?error=invalid_invite", http.StatusSeeOther)
		return
	}

	c.GameID = g.ID
	ownerID := getCookie(r, "cy_user_id")

	if err := s.DB.SaveCharacter(c, ownerID); err != nil {

		log.Printf("Failed to save character %s: %v", c.ID, err)

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return

	}

	setCookie(w, "last_game_invite", g.InviteCode)

	// Broadcast update to real-time clients!
	ws.GlobalHub.Broadcast("game_"+g.ID, "char_update:"+c.ID)
	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)

	http.Redirect(w, r, "/game/"+g.ID, http.StatusSeeOther)
}

func (s *Server) handleUpdateHP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	currStr := r.FormValue("hp_current")
	maxStr := r.FormValue("hp_max")

	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		s.renderError(w, r, "Not found", http.StatusNotFound)
		return
	}

	if val, err := strconv.Atoi(currStr); err == nil {
		c.HP.Current = val
	}
	if val, err := strconv.Atoi(maxStr); err == nil {
		c.HP.Max = val
	}

	ownerID := getCookie(r, "cy_user_id")
	if err := s.DB.SaveCharacter(c, ownerID); err != nil {
		log.Printf("Failed to save character %s: %v", c.ID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Broadcast real-time update to connected WebSockets
	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		s.logAndBroadcastGameEvent(c, "stats", "HP updated to "+strconv.Itoa(c.HP.Current)+"/"+strconv.Itoa(c.HP.Max))
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUpdateGlitches(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	currStr := r.FormValue("glitches_current")
	maxStr := r.FormValue("glitches_max")

	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		s.renderError(w, r, "Not found", http.StatusNotFound)
		return
	}

	if val, err := strconv.Atoi(currStr); err == nil {
		c.Glitches.Current = val
	}
	if val, err := strconv.Atoi(maxStr); err == nil {
		c.Glitches.Max = val
	}

	ownerID := getCookie(r, "cy_user_id")
	if err := s.DB.SaveCharacter(c, ownerID); err != nil {
		log.Printf("Failed to save character %s: %v", c.ID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Broadcast real-time update to connected WebSockets
	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		s.logAndBroadcastGameEvent(c, "stats", "Glitches updated to "+strconv.Itoa(c.Glitches.Current)+"/"+strconv.Itoa(c.Glitches.Max))
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUpdateStat(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	statName := r.FormValue("stat_name")
	currStr := r.FormValue("stat_current")
	maxStr := r.FormValue("stat_max")

	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		s.renderError(w, r, "Not found", http.StatusNotFound)
		return
	}

	if stat, exists := c.Abilities[statName]; exists {
		if val, err := strconv.Atoi(currStr); err == nil {
			stat.Current = val
		}
		if val, err := strconv.Atoi(maxStr); err == nil {
			stat.Max = val
		}
		c.Abilities[statName] = stat

		ownerID := getCookie(r, "cy_user_id")
		if err := s.DB.SaveCharacter(c, ownerID); err != nil {
			log.Printf("Failed to save character %s: %v", c.ID, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
		if c.GameID != "" {
			s.logAndBroadcastGameEvent(c, "stats", "Stat "+statName+" updated to "+r.FormValue("value"))
			ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCreateBlankCharacter(w http.ResponseWriter, r *http.Request) {
	user := s.getUserFromSession(r)
	if user == nil {
		s.renderError(w, r, "Authentication required. Please log in or register an account.", http.StatusUnauthorized)
		return
	}

	c := chargen.CreateBlankCharacter()
	ownerID := user.ID

	if user.Handle != "" {
		c.Handle = user.Handle
	}

	if err := r.ParseForm(); err != nil {

		log.Printf("Failed to parse form: %v", err)

		http.Error(w, "Bad Request", http.StatusBadRequest)

		return

	}
	if gameID := r.FormValue("game_id"); gameID != "" {
		c.GameID = gameID
	}

	if err := s.DB.SaveCharacter(&c, ownerID); err != nil {
		s.renderError(w, r, "Failed to save character", http.StatusInternalServerError)
		return
	}

	setCookie(w, "char_edit_"+c.ID, c.EditCode)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Push-Url", "/character/"+c.ID)
		s.renderCharacterViewWithChar(w, r, &c)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func (s *Server) handleKeepCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		s.renderError(w, r, "Not found", http.StatusNotFound)
		return
	}

	c.IsSaved = true
	ownerID := getCookie(r, "cy_user_id")
	if err := s.DB.SaveCharacter(c, ownerID); err != nil {
		log.Printf("Failed to save character %s: %v", c.ID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	if r.Header.Get("HX-Request") == "true" {
		s.renderCharacterViewWithChar(w, r, c)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func (s *Server) handleUpdateField(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		s.renderError(w, r, "Not found", http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {

		log.Printf("Failed to parse form: %v", err)

		http.Error(w, "Bad Request", http.StatusBadRequest)

		return

	}

	// Handle section multi-field form updates
	if val := r.FormValue("name"); val != "" {
		c.Name = val
	}
	if val := r.FormValue("handle"); val != "" {
		c.Handle = val
	}
	if val := r.FormValue("style"); val != "" {
		c.Style = val
	}
	if val := r.FormValue("feature"); val != "" {
		c.Feature = val
	}
	if val := r.FormValue("want"); val != "" {
		c.Want = val
	}
	if val := r.FormValue("quirk"); val != "" {
		c.Quirk = val
	}
	if val := r.FormValue("obsession"); val != "" {
		c.Obsession = val
	}
	if val := r.FormValue("debt"); val != "" {
		c.Debt = val
	}
	if val := r.FormValue("class_name"); val != "" {
		c.Class.Name = val
	}
	if val := r.FormValue("class_glitch"); val != "" {
		c.Class.Glitch = val
	}
	if val := r.FormValue("class_description"); val != "" {
		c.Class.Description = val
	}
	if val := r.FormValue("class_origin"); val != "" {
		c.Class.Origin = val
	}
	if val := r.FormValue("class_gift"); val != "" {
		c.Class.Gift = val
	}
	if val := r.FormValue("creds"); val != "" {
		if num, err := strconv.Atoi(val); err == nil {
			c.Creds = num
		}
	}

	// Single field legacy fallback
	if field := r.FormValue("field"); field != "" {
		val := r.FormValue("value")
		switch field {
		case "name":
			c.Name = val
		case "handle":
			c.Handle = val
		case "style":
			c.Style = val
		case "feature":
			c.Feature = val
		case "quirk":
			c.Quirk = val
		case "obsession":
			c.Obsession = val
		case "want":
			c.Want = val
		case "debt":
			c.Debt = val
		case "class_name":
			c.Class.Name = val
		case "class_glitch":
			c.Class.Glitch = val
		case "class_description":
			c.Class.Description = val
		case "class_origin":
			c.Class.Origin = val
		case "class_gift":
			c.Class.Gift = val
		case "creds":
			if num, err := strconv.Atoi(val); err == nil {
				c.Creds = num
			}
		}
	}

	ownerID := getCookie(r, "cy_user_id")
	if err := s.DB.SaveCharacter(c, ownerID); err != nil {
		s.renderError(w, r, "Failed to save character", http.StatusInternalServerError)
		return
	}

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		s.logAndBroadcastGameEvent(c, "stats", "Updated character details")
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	if r.Header.Get("HX-Request") == "true" {
		s.renderCharacterViewWithChar(w, r, c)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAddListItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	listType := r.FormValue("type")

	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		s.renderError(w, r, "Not found", http.StatusNotFound)
		return
	}

	switch listType {
	case "weapon":
		name := r.FormValue("name")
		if name == "" {
			name = "New Weapon"
		}
		damage := r.FormValue("damage")
		if damage == "" {
			damage = "d6"
		}
		desc := r.FormValue("description")
		c.Weapons = append(c.Weapons, chargen.Weapon{Name: name, Damage: damage, Description: desc})
	case "armor":
		name := r.FormValue("name")
		if name == "" {
			name = "New Armor"
		}
		tier := r.FormValue("tier")
		if tier == "" {
			tier = "d4"
		}
		reduc := r.FormValue("reduction")
		c.Armor = append(c.Armor, chargen.Armor{Name: name, Tier: tier, Reduction: reduc})
	case "gear":
		item := r.FormValue("item")
		if item != "" {
			c.Gear = append(c.Gear, item)
		}
	case "cybertech":
		item := r.FormValue("item")
		if item != "" {
			c.Cybertech = append(c.Cybertech, item)
		}
	case "app":
		item := r.FormValue("item")
		if item != "" {
			c.Apps = append(c.Apps, item)
		}
	}

	ownerID := getCookie(r, "cy_user_id")
	if err := s.DB.SaveCharacter(c, ownerID); err != nil {
		log.Printf("Failed to save character %s: %v", c.ID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		s.logAndBroadcastGameEvent(c, "inventory", "Added to "+listType)
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	if r.Header.Get("HX-Request") == "true" {
		s.renderCharacterViewWithChar(w, r, c)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func (s *Server) handleDeleteListItem(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderError(w, r, "Invalid form data", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	listType := r.FormValue("type")
	indexStr := r.FormValue("index")
	idx, err := strconv.Atoi(indexStr)

	c, getErr := s.DB.GetCharacter(id)
	if getErr != nil || c == nil || err != nil || idx < 0 {
		s.renderError(w, r, "Invalid request", http.StatusBadRequest)
		return
	}

	sessionCode := getCookie(r, "char_edit_"+c.ID)
	user := s.getUserFromSession(r)
	isOwner := (sessionCode != "" && sessionCode == c.EditCode) || (user != nil && user.ID != "" && c.OwnerID == user.ID)

	isGM := false
	if c.GameID != "" {
		g, _ := s.DB.GetGame(c.GameID)
		if g != nil {
			gameGMCode := getCookie(r, "game_gm_"+g.ID)
			if (gameGMCode != "" && gameGMCode == g.GMCode) || (user != nil && user.ID == g.OwnerID) {
				isGM = true
			}
		}
	}

	if !isOwner && !isGM {
		s.renderError(w, r, "Unauthorized", http.StatusForbidden)
		return
	}


	switch listType {
	case "weapon":
		if idx < len(c.Weapons) {
			c.Weapons = append(c.Weapons[:idx], c.Weapons[idx+1:]...)
		}
	case "armor":
		if idx < len(c.Armor) {
			c.Armor = append(c.Armor[:idx], c.Armor[idx+1:]...)
		}
	case "gear":
		if idx < len(c.Gear) {
			c.Gear = append(c.Gear[:idx], c.Gear[idx+1:]...)
		}
	case "cybertech":
		if idx < len(c.Cybertech) {
			c.Cybertech = append(c.Cybertech[:idx], c.Cybertech[idx+1:]...)
		}
	case "app":
		if idx < len(c.Apps) {
			c.Apps = append(c.Apps[:idx], c.Apps[idx+1:]...)
		}
	}

	ownerID := getCookie(r, "cy_user_id")
	if err := s.DB.SaveCharacter(c, ownerID); err != nil {
		log.Printf("Failed to save character %s: %v", c.ID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		s.logAndBroadcastGameEvent(c, "inventory", "Deleted item from "+listType)
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	if r.Header.Get("HX-Request") == "true" {
		s.renderCharacterViewWithChar(w, r, c)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func (s *Server) handleDeleteCharacter(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderError(w, r, "Invalid form data", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		s.renderError(w, r, "Character not found", http.StatusNotFound)
		return
	}

	sessionCode := getCookie(r, "char_edit_"+c.ID)
	user := s.getUserFromSession(r)
	isOwner := (sessionCode != "" && sessionCode == c.EditCode) || (user != nil && user.ID != "" && c.OwnerID == user.ID)
	
	isGM := false
	if c.GameID != "" {
		g, _ := s.DB.GetGame(c.GameID)
		if g != nil {
			gameGMCode := getCookie(r, "game_gm_"+g.ID)
			if (gameGMCode != "" && gameGMCode == g.GMCode) || (user != nil && user.ID == g.OwnerID) {
				isGM = true
			}
		}
	}

	if !isOwner && !isGM {
		s.renderError(w, r, "Unauthorized: Only the character owner or GM can delete this character", http.StatusForbidden)
		return
	}


	confirmName := r.FormValue("confirm_name")
	if strings.TrimSpace(confirmName) != strings.TrimSpace(c.Name) {
		s.renderError(w, r, fmt.Sprintf("Character name mismatch. You typed '%s', expected '%s'.", confirmName, c.Name), http.StatusBadRequest)
		return
	}

	if err := s.DB.DeleteCharacter(id); err != nil {
		s.renderError(w, r, "Failed to delete character", http.StatusInternalServerError)
		return
	}

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_deleted:"+c.ID)
	if c.GameID != "" {
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleKillCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		s.renderError(w, r, "Character not found", http.StatusNotFound)
		return
	}

	sessionCode := getCookie(r, "char_edit_"+c.ID)
	user := s.getUserFromSession(r)
	isOwner := (sessionCode != "" && sessionCode == c.EditCode) || (user != nil && user.ID != "" && c.OwnerID == user.ID)
	isGM := false
	if c.GameID != "" {
		if g, err := s.DB.GetGame(c.GameID); err == nil && g != nil {
			isGM = getCookie(r, "game_gm_"+g.ID) == g.GMCode
		} else if err != nil {
			log.Printf("Failed to get game %s: %v", c.GameID, err)
		}
	}

	if !isOwner && !isGM {
		s.renderError(w, r, "Unauthorized to mark character flatlined/dead", http.StatusForbidden)
		return
	}

	c.IsDead = true
	c.DiedAt = time.Now()
	deathNote := strings.TrimSpace(r.FormValue("death_note"))
	if deathNote == "" {
		deathNote = "Flatlined in the neon-soaked alleyways of CY."
	}
	c.DeathNote = deathNote

	s.logAndBroadcastGameEvent(c, "death", "has flatlined. "+deathNote)

	ownerID := getCookie(r, "cy_user_id")
	if err := s.DB.SaveCharacter(c, ownerID); err != nil {
		s.renderError(w, r, "Failed to update character", http.StatusInternalServerError)
		return
	}

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	if r.Header.Get("HX-Request") == "true" {
		if r.Header.Get("HX-Target") == "character-sheet-container" {
			s.renderCharacterViewWithChar(w, r, c)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if referer := r.Header.Get("Referer"); referer != "" {
		http.Redirect(w, r, referer, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func (s *Server) handleReviveCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := s.DB.GetCharacter(id)
	if err != nil || c == nil {
		s.renderError(w, r, "Character not found", http.StatusNotFound)
		return
	}

	sessionCode := getCookie(r, "char_edit_"+c.ID)
	user := s.getUserFromSession(r)
	isOwner := (sessionCode != "" && sessionCode == c.EditCode) || (user != nil && user.ID != "" && c.OwnerID == user.ID)
	isGM := false
	if c.GameID != "" {
		if g, err := s.DB.GetGame(c.GameID); err == nil && g != nil {
			isGM = getCookie(r, "game_gm_"+g.ID) == g.GMCode
		} else if err != nil {
			log.Printf("Failed to get game %s: %v", c.GameID, err)
		}
	}

	if !isOwner && !isGM {
		s.renderError(w, r, "Unauthorized to revive character", http.StatusForbidden)
		return
	}

	c.IsDead = false
	c.DiedAt = time.Time{}
	c.DeathNote = ""

	s.logAndBroadcastGameEvent(c, "death", "was revived from the dead.")

	ownerID := getCookie(r, "cy_user_id")
	if err := s.DB.SaveCharacter(c, ownerID); err != nil {
		s.renderError(w, r, "Failed to update character", http.StatusInternalServerError)
		return
	}

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	if r.Header.Get("HX-Request") == "true" {
		if r.Header.Get("HX-Target") == "character-sheet-container" {
			s.renderCharacterViewWithChar(w, r, c)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

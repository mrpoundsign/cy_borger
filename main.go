package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/mrpoundsign/cy_borger/chargen"
	"github.com/mrpoundsign/cy_borger/db"
	"github.com/mrpoundsign/cy_borger/ws"
)

//go:embed templates/*
var templateFS embed.FS

var (
	database  *db.DB
	templates *template.Template
)

func main() {
	var err error
	database, err = db.InitDB("./cy_borger.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	templates, err = template.ParseFS(templateFS, "templates/*.html", "templates/*.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("POST /user/update", handleUpdateUser)
	mux.HandleFunc("POST /character/generate", handleGenerateCharacter)
	mux.HandleFunc("POST /character/create_blank", handleCreateBlankCharacter)
	mux.HandleFunc("GET /character/{id}", handleViewCharacter)
	mux.HandleFunc("POST /character/{id}/auth", handleAuthCharacter)
	mux.HandleFunc("POST /character/{id}/join", handleJoinGame)
	mux.HandleFunc("POST /character/{id}/keep", handleKeepCharacter)
	mux.HandleFunc("PUT /character/{id}/update_hp", handleUpdateHP)
	mux.HandleFunc("PUT /character/{id}/update_glitches", handleUpdateGlitches)
	mux.HandleFunc("PUT /character/{id}/update_stat", handleUpdateStat)
	mux.HandleFunc("PUT /character/{id}/update_field", handleUpdateField)
	mux.HandleFunc("POST /character/{id}/add_item", handleAddListItem)
	mux.HandleFunc("POST /character/{id}/delete_item", handleDeleteListItem)

	mux.HandleFunc("POST /game/create", handleCreateGame)
	mux.HandleFunc("GET /game/{id}", handleViewGame)
	mux.HandleFunc("POST /game/{id}/auth", handleAuthGame)

	// WebSockets (Go stdlib http.Hijacker)
	mux.HandleFunc("GET /ws/game/{id}", handleWSGame)
	mux.HandleFunc("GET /ws/character/{id}", handleWSCharacter)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("⚡ CY_BORG Character Generator running on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
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

// Handlers
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ownerID := getCookie(r, "cy_user_id")
	var myGames []db.Game
	var myChars []chargen.Character
	var user *db.User

	if ownerID != "" {
		user, _ = database.GetUser(ownerID)
		myGames, _ = database.GetGamesByOwner(ownerID)
		myChars, _ = database.GetCharactersByOwner(ownerID)
	}

	data := map[string]interface{}{
		"User":         user,
		"MyGames":      myGames,
		"MyCharacters": myChars,
	}

	_ = templates.ExecuteTemplate(w, "index.html", data)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("user_id")
	username := r.FormValue("username")

	if userID == "" || username == "" {
		http.Error(w, "Missing user_id or username", http.StatusBadRequest)
		return
	}

	// Server-side validation: no spaces/whitespace, and must end with #xxxx
	matched, err := regexp.MatchString(`^[^\s#]+#\d{4}$`, username)
	if err != nil || !matched {
		http.Error(w, "Invalid Operator handle format. Must be Name#XXXX with no spaces.", http.StatusBadRequest)
		return
	}

	u := db.User{ID: userID, Username: username}
	if err := database.SaveUser(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // Send back the "already customized" database lock error
		return
	}

	setCookie(w, "cy_user_id", userID)
	setCookie(w, "cy_username", username)

	w.WriteHeader(http.StatusOK)
}

func handleGenerateCharacter(w http.ResponseWriter, r *http.Request) {
	c := chargen.GenerateCharacter()

	ownerID := getCookie(r, "cy_user_id")
	if ownerID == "" {
		ownerID = "usr_" + chargen.GenerateRandomID(6)
		setCookie(w, "cy_user_id", ownerID)
	}

	// Persist the display name from cookie if available
	username := getCookie(r, "cy_username")
	if username != "" {
		c.Handle = username
	}

	if err := database.SaveCharacter(&c, ownerID); err != nil {
		http.Error(w, "Failed to save character", http.StatusInternalServerError)
		return
	}

	// Set edit session cookie automatically for creator
	setCookie(w, "char_edit_"+c.ID, c.EditCode)

	// If HTMX request, render the character template directly with HX-Push-Url header!
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Push-Url", "/character/"+c.ID)
		renderCharacterViewWithChar(w, r, &c)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func renderCharacterViewWithChar(w http.ResponseWriter, r *http.Request, c *chargen.Character) {
	sessionCode := getCookie(r, "char_edit_"+c.ID)
	canEdit := sessionCode != "" && sessionCode == c.EditCode

	var game *db.Game
	if c.GameID != "" {
		game, _ = database.GetGame(c.GameID)
		if !canEdit && game != nil && getCookie(r, "game_gm_"+game.ID) == game.GMCode {
			canEdit = true
		}
	}

	// Check remembered active game invite if character not in game
	var activeGame *db.Game
	if game != nil {
		activeGame = game
	} else if lastInvite := getCookie(r, "last_game_invite"); lastInvite != "" {
		activeGame, _ = database.GetGameByInviteCode(lastInvite)
	}

	data := map[string]interface{}{
		"Character":  c,
		"CanEdit":    canEdit,
		"Game":       game,
		"ActiveGame": activeGame,
		"IsModal":    r.URL.Query().Get("modal") == "true",
	}

	// Render in modal if requested
	if r.URL.Query().Get("modal") == "true" {
		_ = templates.ExecuteTemplate(w, "character_modal.html", data)
		return
	}

	_ = templates.ExecuteTemplate(w, "character.html", data)
}

func handleViewCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if code := r.URL.Query().Get("code"); code != "" {
		// Set cookie & redirect to clean URL
		setCookie(w, "char_edit_"+id, code)
		http.Redirect(w, r, "/character/"+id, http.StatusSeeOther)
		return
	}

	c, err := database.GetCharacter(id)
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

	renderCharacterViewWithChar(w, r, c)
}

func handleAuthCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	editCode := r.FormValue("edit_code")

	c, err := database.GetCharacter(id)
	if err == nil && c != nil && c.EditCode == editCode {
		setCookie(w, "char_edit_"+id, editCode)
	}

	http.Redirect(w, r, "/character/"+id, http.StatusSeeOther)
}

func handleJoinGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inviteCode := r.FormValue("invite_code")

	c, err := database.GetCharacter(id)
	if err != nil || c == nil {
		http.NotFound(w, r)
		return
	}

	g, err := database.GetGameByInviteCode(inviteCode)
	if err != nil || g == nil {
		http.Redirect(w, r, "/character/"+id+"?error=invalid_invite", http.StatusSeeOther)
		return
	}

	c.GameID = g.ID
	ownerID := getCookie(r, "cy_user_id")

	_ = database.SaveCharacter(c, ownerID)

	setCookie(w, "last_game_invite", g.InviteCode)

	// Broadcast update to real-time clients!
	ws.GlobalHub.Broadcast("game_"+g.ID, "char_update:"+c.ID)
	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)

	http.Redirect(w, r, "/game/"+g.ID, http.StatusSeeOther)
}

func handleUpdateHP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	currStr := r.FormValue("hp_current")
	maxStr := r.FormValue("hp_max")

	c, err := database.GetCharacter(id)
	if err != nil || c == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if val, err := strconv.Atoi(currStr); err == nil {
		c.HP.Current = val
	}
	if val, err := strconv.Atoi(maxStr); err == nil {
		c.HP.Max = val
	}

	ownerID := getCookie(r, "cy_user_id")
	_ = database.SaveCharacter(c, ownerID)

	// Broadcast real-time update to connected WebSockets
	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	w.WriteHeader(http.StatusOK)
}

func handleUpdateGlitches(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	currStr := r.FormValue("glitches_current")
	maxStr := r.FormValue("glitches_max")

	c, err := database.GetCharacter(id)
	if err != nil || c == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if val, err := strconv.Atoi(currStr); err == nil {
		c.Glitches.Current = val
	}
	if val, err := strconv.Atoi(maxStr); err == nil {
		c.Glitches.Max = val
	}

	ownerID := getCookie(r, "cy_user_id")
	_ = database.SaveCharacter(c, ownerID)

	// Broadcast real-time update to connected WebSockets
	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	w.WriteHeader(http.StatusOK)
}

func handleUpdateStat(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	statName := r.FormValue("stat_name")
	currStr := r.FormValue("stat_current")
	maxStr := r.FormValue("stat_max")

	c, err := database.GetCharacter(id)
	if err != nil || c == nil {
		http.Error(w, "Not found", http.StatusNotFound)
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
		_ = database.SaveCharacter(c, ownerID)

		ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
		if c.GameID != "" {
			ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handleCreateGame(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		name = "Unnamed Campaign"
	}

	ownerID := getCookie(r, "cy_user_id")
	if ownerID == "" {
		ownerID = "usr_" + chargen.GenerateRandomID(6)
		setCookie(w, "cy_user_id", ownerID)
	}

	g, err := database.CreateGame(name, ownerID)
	if err != nil {
		http.Error(w, "Failed to create game", http.StatusInternalServerError)
		return
	}

	setCookie(w, "game_gm_"+g.ID, g.GMCode)
	setCookie(w, "last_game_invite", g.InviteCode)

	http.Redirect(w, r, "/game/"+g.ID, http.StatusSeeOther)
}

func handleViewGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if gmCode := r.URL.Query().Get("gm_code"); gmCode != "" {
		setCookie(w, "game_gm_"+id, gmCode)
		http.Redirect(w, r, "/game/"+id, http.StatusSeeOther)
		return
	}

	g, err := database.GetGame(id)
	if err != nil || g == nil {
		http.NotFound(w, r)
		return
	}

	chars, _ := database.GetCharactersForGame(g.ID)

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

	setCookie(w, "last_game_invite", g.InviteCode)

	isGM := getCookie(r, "game_gm_"+g.ID) == g.GMCode

	data := map[string]interface{}{
		"Game":       g,
		"Characters": chars,
		"IsGM":       isGM,
	}

	_ = templates.ExecuteTemplate(w, "game.html", data)
}

func handleAuthGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	gmCode := r.FormValue("gm_code")

	g, err := database.GetGame(id)
	if err == nil && g != nil && g.GMCode == gmCode {
		setCookie(w, "game_gm_"+id, gmCode)
		setCookie(w, "last_game_invite", g.InviteCode)
	}

	http.Redirect(w, r, "/game/"+id, http.StatusSeeOther)
}

// WebSocket Handlers
func handleWSGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ws.ServeWS(w, r, "game_"+id)
}

func handleWSCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ws.ServeWS(w, r, "char_"+id)
}

func handleCreateBlankCharacter(w http.ResponseWriter, r *http.Request) {
	c := chargen.CreateBlankCharacter()

	ownerID := getCookie(r, "cy_user_id")
	if ownerID == "" {
		ownerID = "usr_" + chargen.GenerateRandomID(6)
		setCookie(w, "cy_user_id", ownerID)
	}

	username := getCookie(r, "cy_username")
	if username != "" {
		c.Handle = username
	}

	if err := database.SaveCharacter(&c, ownerID); err != nil {
		http.Error(w, "Failed to save character", http.StatusInternalServerError)
		return
	}

	setCookie(w, "char_edit_"+c.ID, c.EditCode)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Push-Url", "/character/"+c.ID)
		renderCharacterViewWithChar(w, r, &c)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func handleKeepCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := database.GetCharacter(id)
	if err != nil || c == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	c.IsSaved = true
	ownerID := getCookie(r, "cy_user_id")
	_ = database.SaveCharacter(c, ownerID)

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	if r.Header.Get("HX-Request") == "true" {
		renderCharacterViewWithChar(w, r, c)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func handleUpdateField(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	field := r.FormValue("field")
	value := r.FormValue("value")

	c, err := database.GetCharacter(id)
	if err != nil || c == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	switch field {
	case "name":
		c.Name = value
	case "handle":
		c.Handle = value
	case "style":
		c.Style = value
	case "feature":
		c.Feature = value
	case "quirk":
		c.Quirk = value
	case "obsession":
		c.Obsession = value
	case "want":
		c.Want = value
	case "debt":
		c.Debt = value
	case "class_name":
		c.Class.Name = value
	case "class_glitch":
		c.Class.Glitch = value
	case "class_description":
		c.Class.Description = value
	case "class_origin":
		c.Class.Origin = value
	case "class_gift":
		c.Class.Gift = value
	case "creds":
		if val, err := strconv.Atoi(value); err == nil {
			c.Creds = val
		}
	}

	ownerID := getCookie(r, "cy_user_id")
	_ = database.SaveCharacter(c, ownerID)

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	w.WriteHeader(http.StatusOK)
}

func handleAddListItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	listType := r.FormValue("type")

	c, err := database.GetCharacter(id)
	if err != nil || c == nil {
		http.Error(w, "Not found", http.StatusNotFound)
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
	_ = database.SaveCharacter(c, ownerID)

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	if r.Header.Get("HX-Request") == "true" {
		renderCharacterViewWithChar(w, r, c)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func handleDeleteListItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	listType := r.FormValue("type")
	indexStr := r.FormValue("index")
	idx, err := strconv.Atoi(indexStr)

	c, getErr := database.GetCharacter(id)
	if getErr != nil || c == nil || err != nil || idx < 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
	_ = database.SaveCharacter(c, ownerID)

	ws.GlobalHub.Broadcast("char_"+c.ID, "char_update:"+c.ID)
	if c.GameID != "" {
		ws.GlobalHub.Broadcast("game_"+c.GameID, "char_update:"+c.ID)
	}

	if r.Header.Get("HX-Request") == "true" {
		renderCharacterViewWithChar(w, r, c)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

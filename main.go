package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mrpoundsign/cy_borger/chargen"
	"github.com/mrpoundsign/cy_borger/db"
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

	templates, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("POST /character/generate", handleGenerateCharacter)
	mux.HandleFunc("GET /character/{id}", handleViewCharacter)
	mux.HandleFunc("POST /character/{id}/auth", handleAuthCharacter)
	mux.HandleFunc("POST /character/{id}/join", handleJoinGame)
	mux.HandleFunc("PUT /character/{id}/update_hp", handleUpdateHP)
	mux.HandleFunc("PUT /character/{id}/update_glitches", handleUpdateGlitches)

	mux.HandleFunc("POST /game/create", handleCreateGame)
	mux.HandleFunc("GET /game/{id}", handleViewGame)
	mux.HandleFunc("POST /game/{id}/auth", handleAuthGame)

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
	_ = templates.ExecuteTemplate(w, "index.html", nil)
}

func handleGenerateCharacter(w http.ResponseWriter, r *http.Request) {
	c := chargen.GenerateCharacter()
	if err := database.SaveCharacter(&c); err != nil {
		http.Error(w, "Failed to save character", http.StatusInternalServerError)
		return
	}

	// Set edit session cookie automatically for creator
	setCookie(w, "char_edit_"+c.ID, c.EditCode)

	// If HTMX request, render the character template directly with HX-Push-Url header!
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Push-Url", "/character/"+c.ID)
		renderCharacterView(w, r, c.ID)
		return
	}

	http.Redirect(w, r, "/character/"+c.ID, http.StatusSeeOther)
}

func renderCharacterView(w http.ResponseWriter, r *http.Request, id string) {
	c, err := database.GetCharacter(id)
	if err != nil || c == nil {
		http.NotFound(w, r)
		return
	}

	sessionCode := getCookie(r, "char_edit_"+id)
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

	renderCharacterView(w, r, id)
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
	_ = database.SaveCharacter(c)

	// Remember this game invite in cookies for future rolls
	setCookie(w, "last_game_invite", g.InviteCode)

	http.Redirect(w, r, "/game/"+g.ID, http.StatusSeeOther)
}

func handleUpdateHP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	valStr := r.FormValue("hp_current")

	c, err := database.GetCharacter(id)
	if err != nil || c == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if val, err := strconv.Atoi(valStr); err == nil {
		c.HP.Current = val
		_ = database.SaveCharacter(c)
	}

	w.WriteHeader(http.StatusOK)
}

func handleUpdateGlitches(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	valStr := r.FormValue("glitches_current")

	c, err := database.GetCharacter(id)
	if err != nil || c == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if val, err := strconv.Atoi(valStr); err == nil {
		c.Glitches.Current = val
		_ = database.SaveCharacter(c)
	}

	w.WriteHeader(http.StatusOK)
}

func handleCreateGame(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		name = "Unnamed Campaign"
	}

	g, err := database.CreateGame(name)
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

	setCookie(w, "last_game_invite", g.InviteCode)

	chars, _ := database.GetCharactersForGame(g.ID)
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

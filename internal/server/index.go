package server

import (
	"log"
	"net/http"

	"github.com/mrpoundsign/cy_borger/internal/db"
	"github.com/mrpoundsign/cy_borger/pkg/chargen"
	"github.com/mrpoundsign/cy_borger/templates"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		if err := templates.Base("CY_BORGER - 404", nil, templates.NotFound()).Render(r.Context(), w); err != nil {
			log.Printf("Template execution error (404.templ): %v", err)
		}
		return
	}

	user := s.getUserFromSession(r)

	var myGames []db.Game
	var myChars []chargen.Character
	var draftChars []chargen.Character

	if user != nil {
		games, err := s.DB.GetGamesByOwner(user.ID)
		if err != nil {
			log.Printf("Failed to get games for user %s: %v", user.ID, err)
		} else {
			myGames = games
		}

		allChars, err := s.DB.GetCharactersByOwner(user.ID)
		if err != nil {
			log.Printf("Failed to get characters for user %s: %v", user.ID, err)
		}
		for _, c := range allChars {
			if c.IsSaved {
				myChars = append(myChars, c)
			} else {
				draftChars = append(draftChars, c)
			}
		}
	}

	errMsg := r.URL.Query().Get("error")

	if err := templates.Base("CY_BORGER - Home", nil, templates.Index(user, myGames, myChars, draftChars, errMsg)).Render(r.Context(), w); err != nil {
		log.Printf("Template execution error (index.templ): %v", err)
	}
}

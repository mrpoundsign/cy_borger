package server

import (
	"log"
	"net/http"

	"github.com/mrpoundsign/cy_borger/internal/db"
	"github.com/mrpoundsign/cy_borger/pkg/chargen"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		if err := s.Templates.ExecuteTemplate(w, "404.html", nil); err != nil {
			log.Printf("Template execution error (404.html): %v", err)
		}
		return
	}

	user := s.getUserFromSession(r)

	var myGames []db.Game
	var myChars []chargen.Character
	var draftChars []chargen.Character

	if user != nil {
		myGames, _ = s.DB.GetGamesByOwner(user.ID)
		allChars, _ := s.DB.GetCharactersByOwner(user.ID)
		for _, c := range allChars {
			if c.IsSaved {
				myChars = append(myChars, c)
			} else {
				draftChars = append(draftChars, c)
			}
		}
	}

	errMsg := r.URL.Query().Get("error")

	data := map[string]interface{}{
		"User":            user,
		"MyGames":         myGames,
		"MyCharacters":    myChars,
		"DraftCharacters": draftChars,
		"ErrorMessage":    errMsg,
	}

	if err := s.Templates.ExecuteTemplate(w, "index.html", data); err != nil {

		log.Printf("Template execution error (index.html): %v", err)

	}
}

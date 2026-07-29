package server

import (
	"log"
	"net/http"

	"github.com/mrpoundsign/cy_borger/internal/templates"
)

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	if err := templates.About().Render(r.Context(), w); err != nil {
		log.Printf("Template execution error (about.templ): %v", err)
	}
}

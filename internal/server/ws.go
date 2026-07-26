package server

import (
	"net/http"

	"github.com/mrpoundsign/cy_borger/internal/ws"
)

// WebSocket Handlers
func (s *Server) handleWSGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ws.ServeWS(w, r, "game_"+id)
}

func (s *Server) handleWSCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ws.ServeWS(w, r, "char_"+id)
}

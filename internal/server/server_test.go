package server

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrpoundsign/cy_borger/internal/db"
)

func setupTestServer(t *testing.T) (*Server, *http.ServeMux) {
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize memory DB: %v", err)
	}

	tmpls := make(map[string]*template.Template)
	dummy := template.Must(template.New("test").Parse(`{{define "base"}}INDEX{{end}}`))
	tmpls["index.html"] = dummy
	srv := NewServer(database)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	return srv, mux
}

func TestIndex(t *testing.T) {
	_, mux := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "CY_BORGER") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mrpoundsign/cy_borger/internal/db"
	"github.com/mrpoundsign/cy_borger/internal/server"
	"github.com/mrpoundsign/cy_borger/static"
	"github.com/mrpoundsign/cy_borger/templates"
)

func main() {
	var err error
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./cy_borger.db"
	}
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	tmpl, err := template.ParseFS(templates.FS, "*.html", "*.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	appServer := server.NewServer(database, tmpl)

	mux := http.NewServeMux()

	// Register all server routes
	appServer.RegisterRoutes(mux)

	// Static Files & Favicon (Served from embedded FS)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		content, err := static.FS.ReadFile("favicon.svg")
		if err != nil {
			http.Error(w, "Failed to load favicon", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		if _, err := w.Write(content); err != nil {
			log.Printf("Error writing favicon: %v", err)
		}
	})
	mux.HandleFunc("GET /favicon.png", func(w http.ResponseWriter, r *http.Request) {
		content, err := static.FS.ReadFile("favicon.png")
		if err != nil {
			http.Error(w, "Failed to load favicon", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		if _, err := w.Write(content); err != nil {
			log.Printf("Error writing favicon: %v", err)
		}
	})
	mux.HandleFunc("GET /favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		content, err := static.FS.ReadFile("favicon.svg")
		if err != nil {
			http.Error(w, "Failed to load favicon", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		if _, err := w.Write(content); err != nil {
			log.Printf("Error writing favicon: %v", err)
		}
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("⚡ CY_BORGER Character Generator running on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}

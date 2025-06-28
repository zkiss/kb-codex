package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	go_openai "github.com/sashabaranov/go-openai"
	"github.com/zkiss/kb-codex/internal/config"
	"github.com/zkiss/kb-codex/internal/db"
	"github.com/zkiss/kb-codex/internal/handlers"
)

func main() {
	cfg := config.Load()

	dbConn, err := db.ConnectAndMigrate(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer dbConn.Close()

	authHandler := handlers.NewAuthHandler(dbConn, []byte(cfg.JWTSecret))

	openaiClient := go_openai.NewClient(cfg.OpenAIAPIKey)
	kbHandler := handlers.NewKBHandler(dbConn, openaiClient)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/api/register", authHandler.Register)
	r.Post("/api/login", authHandler.Login)

	r.Get("/api/kbs", kbHandler.ListKB)
	r.Post("/api/kbs", kbHandler.CreateKB)
	r.Get("/api/kbs/{kbID}/files", kbHandler.ListFiles)
	r.Post("/api/kbs/{kbID}/files", kbHandler.UploadFile)
	r.Post("/api/kbs/{kbID}/ask", kbHandler.AskQuestion)

	// Serve SPA static files
	r.Handle("/*", spaHandler("static", "index.html"))

	log.Printf("Starting server on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

// spaHandler serves static files and falls back to the index file for unknown paths.
func spaHandler(staticDir, indexFile string) http.HandlerFunc {
	fs := http.Dir(staticDir)
	fileServer := http.FileServer(fs)
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		full := filepath.Join(staticDir, path)
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticDir, indexFile))
	}
}

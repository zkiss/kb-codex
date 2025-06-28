package main

import (
	"log"
	"net/http"

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

	// Serve static assets like app.js
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Post("/api/register", authHandler.Register)
	r.Post("/api/login", authHandler.Login)

	r.Get("/api/kbs", kbHandler.ListKB)
	r.Post("/api/kbs", kbHandler.CreateKB)
	r.Get("/api/kbs/{kbID}/files", kbHandler.ListFiles)
	r.Post("/api/kbs/{kbID}/files", kbHandler.UploadFile)
	r.Post("/api/kbs/{kbID}/ask", kbHandler.AskQuestion)

	// Serve SPA
	r.Get("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	log.Printf("Starting server on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

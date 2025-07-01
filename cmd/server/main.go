package main

import (
	"log"
	"net/http"

	go_openai "github.com/sashabaranov/go-openai"

	"github.com/zkiss/kb-codex/internal/app"
	"github.com/zkiss/kb-codex/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	openaiClient := go_openai.NewClient(cfg.OpenAIAPIKey)
	dbConn, router, err := app.New(app.Dependencies{
		DatabaseURL: cfg.DatabaseURL,
		JWTSecret:   []byte(cfg.JWTSecret),
		AIClient:    openaiClient,
	})
	if err != nil {
		log.Fatalf("could not set up app: %v", err)
	}
	defer dbConn.Close()

	log.Printf("Starting server on %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

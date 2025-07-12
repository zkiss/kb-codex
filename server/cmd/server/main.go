package main

import (
	"fmt"
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
	appInstance, err := app.New(cfg, openaiClient)
	if err != nil {
		log.Fatalf("could not set up app: %v", err)
	}
	defer appInstance.Close()

	log.Printf("Starting server on %d", cfg.Port)
	err = appInstance.Listen(func(port uint16, router http.Handler) error {
		return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	})
	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

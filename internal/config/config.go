package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds configuration settings for the application.
type Config struct {
	Addr         string
	DatabaseURL  string
	JWTSecret    string
	OpenAIAPIKey string
}

// Load reads configuration from flags and environment variables.
func Load() *Config {
	addr := flag.String("addr", ":8080", "HTTP network address")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("WARNING: DATABASE_URL is not set. Using default postgres connection")
		dbURL = "postgres://demo:demo_pw@localhost:5432/postgres?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Println("WARNING: JWT_SECRET is not set. Using insecure default secret")
		jwtSecret = "secret"
	}

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		fmt.Println("WARNING: OPENAI_API_KEY is not set. Embedding requests will fail.")
	}

	return &Config{
		Addr:         *addr,
		DatabaseURL:  dbURL,
		JWTSecret:    jwtSecret,
		OpenAIAPIKey: openAIKey,
	}
}

package config

import (
	"fmt"
	"os"
)

// Config holds configuration settings for the application.
type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	OpenAIAPIKey string
}

// Load reads configuration from flags and environment variables.
func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

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
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required but not set")
	}

	return &Config{
		Port:         port,
		DatabaseURL:  dbURL,
		JWTSecret:    jwtSecret,
		OpenAIAPIKey: openAIKey,
	}, nil
}

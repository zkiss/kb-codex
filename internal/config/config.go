package config

import (
	"fmt"
	"os"
)

// Config holds configuration settings for the application.
type Config struct {
	Port         uint16
	DatabaseURL  string
	JWTSecret    []byte
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

	// Convert port from string to uint16
	var portUint uint16
	_, err := fmt.Sscanf(port, "%d", &portUint)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT value: %v", err)
	}

	return &Config{
		Port:         portUint,
		DatabaseURL:  dbURL,
		JWTSecret:    []byte(jwtSecret),
		OpenAIAPIKey: openAIKey,
	}, nil
}

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
		return nil, fmt.Errorf("DATABASE_URL environment variable is required but not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required but not set")
	}

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required but not set")
	}

	// Convert port from string to uint16
	var portUint uint16
	_, err := fmt.Sscanf(port, "%d", &portUint)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT value '%s': %v", port, err)
	}

	return &Config{
		Port:         portUint,
		DatabaseURL:  dbURL,
		JWTSecret:    []byte(jwtSecret),
		OpenAIAPIKey: openAIKey,
	}, nil
}

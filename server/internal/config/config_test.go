package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")
	os.Setenv("OPENAI_API_KEY", "openai-key-123")

	cfg, _ := Load()

	assert.NotEmpty(t, cfg.DatabaseURL)
	assert.NotEmpty(t, cfg.JWTSecret)
}

func TestLoadFromEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/dbname")
	os.Setenv("JWT_SECRET", "supersecretjwt")
	os.Setenv("OPENAI_API_KEY", "openai-key-123")

	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("OPENAI_API_KEY")

	cfg, _ := Load()

	assert.Equal(t, "postgres://user:pass@localhost:5432/dbname", cfg.DatabaseURL)
	assert.Equal(t, []byte("supersecretjwt"), cfg.JWTSecret)
	assert.Equal(t, "openai-key-123", cfg.OpenAIAPIKey)
}

func TestLoadMissingOpenAIKey(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/dbname")
	os.Setenv("JWT_SECRET", "supersecretjwt")
	os.Unsetenv("OPENAI_API_KEY")

	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("OPENAI_API_KEY")

	_, err := Load()
	assert.Error(t, err, "expected error when OPENAI_API_KEY is missing")
	assert.Contains(t, err.Error(), "OPENAI_API_KEY")
}

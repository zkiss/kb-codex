package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaults(t *testing.T) {
	os.Setenv("JWT_SECRET", "some-jwt-secret")
	os.Setenv("OPENAI_API_KEY", "openai-key-123")
	os.Setenv("DATABASE_URL", "postgres://some-db-url")

	cfg, _ := Load()

	assert.NotEmpty(t, cfg.Port)
}

func TestLoadFromEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/dbname")
	os.Setenv("JWT_SECRET", "supersecretjwt")
	os.Setenv("OPENAI_API_KEY", "openai-key-123")
	os.Setenv("PORT", "1234")

	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("PORT")

	cfg, _ := Load()

	assert.Equal(t, "postgres://user:pass@localhost:5432/dbname", cfg.DatabaseURL)
	assert.Equal(t, []byte("supersecretjwt"), cfg.JWTSecret)
	assert.Equal(t, "openai-key-123", cfg.OpenAIAPIKey)
	assert.Equal(t, uint16(1234), cfg.Port)
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

func TestLoadMissingDatabaseURL(t *testing.T) {
	os.Setenv("JWT_SECRET", "supersecretjwt")
	os.Setenv("OPENAI_API_KEY", "openai-key-123")

	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("OPENAI_API_KEY")

	_, err := Load()
	assert.Error(t, err, "expected error when DATABASE_URL is missing")
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoadMissingJWTKey(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/dbname")
	os.Unsetenv("JWT_SECRET")
	os.Setenv("OPENAI_API_KEY", "openai-key-123")

	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("OPENAI_API_KEY")

	_, err := Load()
	assert.Error(t, err, "expected error when JWT_SECRET is missing")
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

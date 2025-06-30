package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("OPENAI_API_KEY")

	cfg := Load()

	assert.NotEmpty(t, cfg.DatabaseURL)
	assert.NotEmpty(t, cfg.JWTSecret)
}

package testutil

import (
	"os"
	"strings"
	"testing"
)

func RequireDocker(t *testing.T) {
	t.Helper()
	if v := os.Getenv("SKIP_CONTAINER_TESTS"); strings.ToLower(v) == "1" || strings.ToLower(v) == "true" {
		t.Skip("skipping container-based tests because SKIP_CONTAINER_TESTS is set")
	}
}
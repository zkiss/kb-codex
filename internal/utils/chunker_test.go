package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunkText(t *testing.T) {
	text := "this is a test of the emergency broadcast system"
	chunks := ChunkText(text, 10)
	expected := []string{"this is a", "test of", "the", "emergency", "broadcast", "system"}
	assert.Equal(t, expected, chunks)
}

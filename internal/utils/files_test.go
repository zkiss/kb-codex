package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugifyFileName(t *testing.T) {
	assert.Equal(t, "hello-world", SlugifyFileName("Hello World"))
	assert.Equal(t, "my-file-123-txt", SlugifyFileName("my_file 123!!.txt"))
}

func TestRandomString(t *testing.T) {
	s := RandomString(6)
	assert.Equal(t, 6, len(s))
}

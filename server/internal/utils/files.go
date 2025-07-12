package utils

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
var multiDash = regexp.MustCompile(`-+`)

// SlugifyFileName converts a filename to a lowercase kebab-case string
// containing only alphanumeric characters and dashes.
func SlugifyFileName(name string) string {
	name = strings.ToLower(name)
	name = nonAlnum.ReplaceAllString(name, "-")
	name = multiDash.ReplaceAllString(name, "-")
	return strings.Trim(name, "-")
}

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandomString returns a random alphanumeric string of length n.
func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

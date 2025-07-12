package utils

import (
	"strings"
)

// ChunkText splits text into chunks of up to maxLen characters, breaking on word boundaries.
func ChunkText(text string, maxLen int) []string {
	var chunks []string
	words := strings.Fields(text)
	var current strings.Builder

	for _, w := range words {
		if current.Len()+len(w)+1 > maxLen && current.Len() > 0 {
			chunks = append(chunks, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(w)
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}
	return chunks
}

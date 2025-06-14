package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	go_openai "github.com/sashabaranov/go-openai"

	"github.com/zkiss/kb-codex/internal/utils"
)

// KBHandler provides endpoints for managing knowledge bases and file uploads.
type KBHandler struct {
	DB     *sql.DB
	OpenAI *go_openai.Client
}

// NewKBHandler constructs a KBHandler instance.
func NewKBHandler(db *sql.DB, openaiClient *go_openai.Client) *KBHandler {
	return &KBHandler{DB: db, OpenAI: openaiClient}
}

// createKBRequest represents the JSON payload for creating a knowledge base.
type createKBRequest struct {
	Name string `json:"name"`
}

// KB represents a knowledge base.
type KB struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateKB handles POST /api/kbs
func (h *KBHandler) CreateKB(w http.ResponseWriter, r *http.Request) {
	var req createKBRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	var id int64
	var createdAt time.Time
	err := h.DB.QueryRow(
		`INSERT INTO knowledge_bases(name) VALUES ($1) RETURNING id, created_at`,
		req.Name,
	).Scan(&id, &createdAt)
	if err != nil {
		http.Error(w, "could not create knowledge base: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(KB{ID: id, Name: req.Name, CreatedAt: createdAt})
}

// ListKB handles GET /api/kbs
func (h *KBHandler) ListKB(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT id, name, created_at FROM knowledge_bases ORDER BY id`)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []KB
	for rows.Next() {
		var kb KB
		if err := rows.Scan(&kb.ID, &kb.Name, &kb.CreatedAt); err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		list = append(list, kb)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// ListFiles handles GET /api/kbs/{kbID}/files
func (h *KBHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	kbIDStr := chi.URLParam(r, "kbID")
	kbID, err := strconv.ParseInt(kbIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid kb ID", http.StatusBadRequest)
		return
	}
	rows, err := h.DB.Query(
		`SELECT DISTINCT file_name FROM chunks WHERE kb_id = $1 ORDER BY file_name`,
		kbID,
	)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var fname string
		if err := rows.Scan(&fname); err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		files = append(files, fname)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// UploadFile handles POST /api/kbs/{kbID}/files (multipart file upload)
func (h *KBHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	kbIDStr := chi.URLParam(r, "kbID")
	kbID, err := strconv.ParseInt(kbIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid kb ID", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".txt" && ext != ".md" {
		http.Error(w, "only .txt and .md files are supported", http.StatusBadRequest)
		return
	}
	contentBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "could not read file", http.StatusInternalServerError)
		return
	}
	content := string(contentBytes)
	chunks := utils.ChunkText(content, 1000)
	ctx := r.Context()
	for idx, chunk := range chunks {
		embReq := go_openai.EmbeddingRequest{
			Model: go_openai.AdaEmbeddingV2,
			Input: []string{chunk},
		}
		embResp, err := h.OpenAI.CreateEmbeddings(ctx, embReq)
		if err != nil {
			http.Error(w, "embedding failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		vec := embResp.Data[0].Embedding
		parts := make([]string, len(vec))
		for i, v := range vec {
			parts[i] = fmt.Sprintf("%f", v)
		}
		arrLit := "[" + strings.Join(parts, ",") + "]"
		_, err = h.DB.Exec(
			`INSERT INTO chunks(kb_id, file_name, chunk_index, content, embedding) VALUES($1,$2,$3,$4,$5::vector)`,
			kbID, header.Filename, idx, chunk, arrLit,
		)
		if err != nil {
			http.Error(w, "could not save chunk: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"chunks": len(chunks)})
}

// questionRequest represents a question about a knowledge base.
type questionRequest struct {
	Question string `json:"question"`
}

// questionResponse represents the answer returned to the client.
type questionResponse struct {
	Answer string `json:"answer"`
}

// AskQuestion handles POST /api/kbs/{kbID}/ask
func (h *KBHandler) AskQuestion(w http.ResponseWriter, r *http.Request) {
	kbIDStr := chi.URLParam(r, "kbID")
	kbID, err := strconv.ParseInt(kbIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid kb ID", http.StatusBadRequest)
		return
	}
	var req questionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	embReq := go_openai.EmbeddingRequest{
		Model: go_openai.AdaEmbeddingV2,
		Input: []string{req.Question},
	}
	embResp, err := h.OpenAI.CreateEmbeddings(ctx, embReq)
	if err != nil {
		http.Error(w, "embedding failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	vec := embResp.Data[0].Embedding
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%f", v)
	}
	arrLit := "[" + strings.Join(parts, ",") + "]"

	rows, err := h.DB.QueryContext(ctx,
		`SELECT content FROM chunks WHERE kb_id=$1 ORDER BY embedding <-> $2::vector LIMIT 5`,
		kbID, arrLit,
	)
	if err != nil {
		http.Error(w, "search failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var contextParts []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			http.Error(w, "search failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		contextParts = append(contextParts, c)
	}

	prompt := fmt.Sprintf("Answer the question based on the following context:\n\n%s\n\nQuestion: %s",
		strings.Join(contextParts, "\n---\n"), req.Question)

	chatReq := go_openai.ChatCompletionRequest{
		Model: go_openai.GPT3Dot5Turbo,
		Messages: []go_openai.ChatCompletionMessage{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: prompt},
		},
	}
	chatResp, err := h.OpenAI.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		http.Error(w, "openai failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	answer := chatResp.Choices[0].Message.Content

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questionResponse{Answer: answer})
}

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	go_openai "github.com/sashabaranov/go-openai"

	"bytes"

	pdf "github.com/ledongthuc/pdf"
	"github.com/zkiss/kb-codex/internal/utils"
)

// KBHandler provides endpoints for managing knowledge bases and file uploads.
// AIClient defines the subset of the OpenAI client used by the handler. It
// allows injecting a fake implementation in tests so no real network calls are
// made.
type AIClient interface {
	CreateEmbeddings(ctx context.Context, req go_openai.EmbeddingRequestConverter) (go_openai.EmbeddingResponse, error)
	CreateChatCompletion(ctx context.Context, req go_openai.ChatCompletionRequest) (go_openai.ChatCompletionResponse, error)
}

// KBHandler provides endpoints for managing knowledge bases and file uploads.
type KBHandler struct {
	DB     *sql.DB
	OpenAI AIClient
}

// NewKBHandler constructs a KBHandler instance.
func NewKBHandler(db *sql.DB, openaiClient AIClient) *KBHandler {
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
		`SELECT file_name, lookup_name FROM files WHERE kb_id = $1 ORDER BY file_name`,
		kbID,
	)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type fileEntry struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	var files []fileEntry
	for rows.Next() {
		var fname, slug string
		if err := rows.Scan(&fname, &slug); err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		files = append(files, fileEntry{Name: fname, Slug: slug})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// GetFile handles GET /api/kbs/{kbID}/files/{slug}
func (h *KBHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	kbIDStr := chi.URLParam(r, "kbID")
	kbID, err := strconv.ParseInt(kbIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid kb ID", http.StatusBadRequest)
		return
	}
	slug := chi.URLParam(r, "slug")
	var content []byte
	var mimeType string
	var fileName string
	err = h.DB.QueryRow(`SELECT file_name, content, mime_type FROM files WHERE kb_id=$1 AND lookup_name=$2`, kbID, slug).Scan(&fileName, &content, &mimeType)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}
	if mimeType == "" {
		ext := strings.ToLower(filepath.Ext(fileName))
		switch ext {
		case ".md":
			mimeType = "text/markdown; charset=utf-8"
		default:
			mimeType = "text/plain; charset=utf-8"
		}
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", mimeType)
	w.Write(content)
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
	if ext != ".txt" && ext != ".md" && ext != ".pdf" {
		http.Error(w, "only .txt, .md, and .pdf files are supported", http.StatusBadRequest)
		return
	}
	contentBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "could not read file", http.StatusInternalServerError)
		return
	}
	var contentStr string
	if ext == ".pdf" {
		// Extract text from PDF
		pdfText, err := extractTextFromPDF(contentBytes)
		if err != nil {
			http.Error(w, "could not extract text from PDF: "+err.Error(), http.StatusInternalServerError)
			return
		}
		contentStr = pdfText
	} else {
		contentStr = string(contentBytes)
	}
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mime.TypeByExtension(ext)
	}

	lookup := utils.SlugifyFileName(header.Filename)
	if len(lookup) > 50 {
		lookup = lookup[:50]
	}
	var exists int
	err = h.DB.QueryRow(`SELECT 1 FROM files WHERE kb_id=$1 AND lookup_name=$2`, kbID, lookup).Scan(&exists)
	if err != sql.ErrNoRows && err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if err != sql.ErrNoRows {
		base := lookup
		if len(base) > 43 {
			base = base[:43]
		}
		lookup = base + "-" + utils.RandomString(6)
	}

	_, err = h.DB.Exec(`INSERT INTO files(kb_id, file_name, lookup_name, mime_type, content, created_at) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT (kb_id, lookup_name) DO UPDATE SET file_name=EXCLUDED.file_name, mime_type=EXCLUDED.mime_type, content=EXCLUDED.content, created_at=EXCLUDED.created_at`,
		kbID, header.Filename, lookup, mimeType, contentBytes, time.Now())
	if err != nil {
		http.Error(w, "could not store file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	chunks := utils.ChunkText(contentStr, 1000)
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
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type questionRequest struct {
	Question string        `json:"question"`
	History  []chatMessage `json:"history"`
}

// questionResponse represents the answer returned to the client.
type questionChunk struct {
	FileName string `json:"file_name"`
	Index    int    `json:"index"`
	Content  string `json:"content"`
}

type questionResponse struct {
	Answer string          `json:"answer"`
	Chunks []questionChunk `json:"chunks"`
}

func rewriteQuestion(ctx context.Context, ai AIClient, history []chatMessage, q string) (string, error) {
	messages := []go_openai.ChatCompletionMessage{
		{Role: "system", Content: "Rewrite the user's question to be a standalone question using the conversation history."},
	}
	for _, m := range history {
		messages = append(messages, go_openai.ChatCompletionMessage{Role: m.Role, Content: m.Content})
	}
	messages = append(messages, go_openai.ChatCompletionMessage{Role: "user", Content: q})
	resp, err := ai.CreateChatCompletion(ctx, go_openai.ChatCompletionRequest{
		Model:    go_openai.GPT3Dot5Turbo,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
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
	q := req.Question
	if len(req.History) > 0 {
		q, err = rewriteQuestion(ctx, h.OpenAI, req.History, req.Question)
		if err != nil {
			http.Error(w, "question rewrite failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	embReq := go_openai.EmbeddingRequest{
		Model: go_openai.AdaEmbeddingV2,
		Input: []string{q},
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
		`SELECT file_name, chunk_index, content FROM chunks WHERE kb_id=$1 ORDER BY embedding <-> $2::vector LIMIT 5`,
		kbID, arrLit,
	)
	if err != nil {
		http.Error(w, "search failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var contextParts []string
	var chunks []questionChunk
	for rows.Next() {
		var c questionChunk
		if err := rows.Scan(&c.FileName, &c.Index, &c.Content); err != nil {
			http.Error(w, "search failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		contextParts = append(contextParts, c.Content)
		chunks = append(chunks, c)
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
	json.NewEncoder(w).Encode(questionResponse{Answer: answer, Chunks: chunks})
}

// extractTextFromPDF extracts text from a PDF file given as []byte.
func extractTextFromPDF(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	// Iterate through all pages
	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		buf.WriteString(content)
		buf.WriteString("\n")
	}
	return buf.String(), nil
}

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	go_openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/zkiss/kb-codex/internal/testutil"
	"github.com/zkiss/kb-codex/internal/utils"
)

// helper to start a postgres container with a simplified schema
func setupVectorDB(t *testing.T, dim int) (*postgres.PostgresContainer, *sql.DB) {
	t.Helper()
	testutil.RequireDocker(t)
	ctx := context.Background()
	pg, err := postgres.Run(ctx, "pgvector/pgvector:pg17",
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("demo"),
		postgres.WithPassword("demo_pw"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("starting postgres: %v", err)
	}
	dbURL, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		pg.Terminate(ctx)
		t.Fatalf("conn string: %v", err)
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		pg.Terminate(ctx)
		t.Fatalf("open db: %v", err)
	}
	// Enable pgvector extension
	if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS vector`); err != nil {
		pg.Terminate(ctx)
		t.Fatalf("enable pgvector: %v", err)
	}
	schema := fmt.Sprintf(`CREATE TABLE users(id SERIAL PRIMARY KEY, email TEXT UNIQUE, password_hash TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
CREATE TABLE knowledge_bases(id SERIAL PRIMARY KEY, name TEXT, user_id INTEGER REFERENCES users(id));
CREATE TABLE chunks(id SERIAL PRIMARY KEY, kb_id INTEGER, file_name TEXT, chunk_index INTEGER, content TEXT, embedding VECTOR(%d));`, dim)
	if _, err := db.Exec(schema); err != nil {
		pg.Terminate(ctx)
		t.Fatalf("create tables: %v", err)
	}
	return pg, db
}

type recordingAI struct {
	emb           []float32
	lastPrompt    string
	lastEmbInput  string
	rewriteCalled bool
}

func (r *recordingAI) CreateEmbeddings(ctx context.Context, req go_openai.EmbeddingRequestConverter) (go_openai.EmbeddingResponse, error) {
	conv := req.Convert()
	if in, ok := conv.Input.([]string); ok && len(in) > 0 {
		r.lastEmbInput = in[0]
	}
	return go_openai.EmbeddingResponse{Data: []go_openai.Embedding{{Embedding: r.emb}}}, nil
}

func (r *recordingAI) CreateChatCompletion(ctx context.Context, req go_openai.ChatCompletionRequest) (go_openai.ChatCompletionResponse, error) {
	if strings.Contains(req.Messages[0].Content, "Rewrite") {
		r.rewriteCalled = true
		return go_openai.ChatCompletionResponse{Choices: []go_openai.ChatCompletionChoice{{Message: go_openai.ChatCompletionMessage{Content: "rewritten"}}}}, nil
	}
	r.lastPrompt = req.Messages[len(req.Messages)-1].Content
	return go_openai.ChatCompletionResponse{Choices: []go_openai.ChatCompletionChoice{{Message: go_openai.ChatCompletionMessage{Content: "answer"}}}}, nil
}

func toArrayLit(vec []float32) string {
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%f", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func TestAskQuestionComponent(t *testing.T) {
	pg, db := setupVectorDB(t, 3)
	defer pg.Terminate(context.Background())
	defer db.Close()

	// Create a test user
	var userID int64
	err := db.QueryRow(`INSERT INTO users(email, password_hash, created_at, updated_at) VALUES('test@example.com', 'hash', NOW(), NOW()) RETURNING id`).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	var kbID int64
	err = db.QueryRow(`INSERT INTO knowledge_bases(name, user_id) VALUES('kb1', $1) RETURNING id`, userID).Scan(&kbID)
	if err != nil {
		t.Fatalf("insert kb: %v", err)
	}
	// insert chunks
	_, err = db.Exec(`INSERT INTO chunks(kb_id,file_name,chunk_index,content,embedding) VALUES($1,'f.txt',0,'alpha',$2::vector)`, kbID, toArrayLit([]float32{1, 0, 0}))
	if err != nil {
		t.Fatalf("insert chunk: %v", err)
	}
	_, err = db.Exec(`INSERT INTO chunks(kb_id,file_name,chunk_index,content,embedding) VALUES($1,'f.txt',1,'bravo',$2::vector)`, kbID, toArrayLit([]float32{0, 1, 0}))
	if err != nil {
		t.Fatalf("insert chunk: %v", err)
	}

	ai := &recordingAI{emb: []float32{1, 0, 0}}
	h := NewKBHandler(db, ai)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/kbs/%d/ask", kbID), strings.NewReader(`{"question":"q"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("kbID", fmt.Sprint(kbID))
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.AskQuestion(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, ai.lastPrompt, "alpha")
	var resp map[string]any
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "answer", resp["answer"])
}

func TestAskQuestionFollowup(t *testing.T) {
	pg, db := setupVectorDB(t, 3)
	defer pg.Terminate(context.Background())
	defer db.Close()

	// Create a test user
	var userID int64
	err := db.QueryRow(`INSERT INTO users(email, password_hash, created_at, updated_at) VALUES('test2@example.com', 'hash', NOW(), NOW()) RETURNING id`).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	var kbID int64
	err = db.QueryRow(`INSERT INTO knowledge_bases(name, user_id) VALUES('kb1', $1) RETURNING id`, userID).Scan(&kbID)
	if err != nil {
		t.Fatalf("insert kb: %v", err)
	}
	_, err = db.Exec(`INSERT INTO chunks(kb_id,file_name,chunk_index,content,embedding) VALUES($1,'f.txt',0,'alpha',$2::vector)`, kbID, toArrayLit([]float32{1, 0, 0}))
	if err != nil {
		t.Fatalf("insert chunk: %v", err)
	}

	ai := &recordingAI{emb: []float32{1, 0, 0}}
	h := NewKBHandler(db, ai)

	body := `{"question":"follow?","history":[{"role":"user","content":"first"},{"role":"assistant","content":"a1"}]}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/kbs/%d/ask", kbID), strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("kbID", fmt.Sprint(kbID))
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.AskQuestion(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, ai.rewriteCalled, "rewrite should be called")
	assert.Equal(t, "rewritten", ai.lastEmbInput)
}

func TestExtractTextFromPDF(t *testing.T) {
	// Minimal PDF with the text 'Hello PDF'
	pdfBytes, err := os.ReadFile("testdata/pdf_test.pdf")
	if err != nil {
		t.Fatalf("failed to read test PDF: %v", err)
	}
	text, err := extractTextFromPDF(pdfBytes)
	assert.NoError(t, err)
	assert.Contains(t, text, "pdf test")
}

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	go_openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// helper to start a postgres container with a simplified schema
func setupVectorDB(t *testing.T, dim int) (*postgres.PostgresContainer, *sql.DB) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}
	ctx := context.Background()
	pg, err := postgres.Run(ctx, "pgvector/pgvector:pg17",
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("demo"),
		postgres.WithPassword("demo_pw"),
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
	schema := fmt.Sprintf(`CREATE TABLE knowledge_bases(id SERIAL PRIMARY KEY, name TEXT);
CREATE TABLE chunks(id SERIAL PRIMARY KEY, kb_id INTEGER, file_name TEXT, chunk_index INTEGER, content TEXT, embedding VECTOR(%d));`, dim)
	if _, err := db.Exec(schema); err != nil {
		pg.Terminate(ctx)
		t.Fatalf("create tables: %v", err)
	}
	return pg, db
}

type recordingAI struct {
	emb        []float32
	lastPrompt string
}

func (r *recordingAI) CreateEmbeddings(ctx context.Context, req go_openai.EmbeddingRequestConverter) (go_openai.EmbeddingResponse, error) {
	return go_openai.EmbeddingResponse{Data: []go_openai.Embedding{{Embedding: r.emb}}}, nil
}

func (r *recordingAI) CreateChatCompletion(ctx context.Context, req go_openai.ChatCompletionRequest) (go_openai.ChatCompletionResponse, error) {
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

	var kbID int64
	err := db.QueryRow(`INSERT INTO knowledge_bases(name) VALUES('kb1') RETURNING id`).Scan(&kbID)
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
	rr := httptest.NewRecorder()
	h.AskQuestion(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, ai.lastPrompt, "alpha")
	var resp map[string]any
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "answer", resp["answer"])
}

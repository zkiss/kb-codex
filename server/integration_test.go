package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	go_openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"golang.org/x/crypto/bcrypt"

	"github.com/zkiss/kb-codex/internal/app"
	"github.com/zkiss/kb-codex/internal/config"
	"github.com/zkiss/kb-codex/internal/handlers"
	"github.com/zkiss/kb-codex/internal/testutil"
)

type testApp struct {
	srv *httptest.Server
	ai  *fakeAI
}

func setupApp(t *testing.T) *testApp {
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
	t.Cleanup(func() { pg.Terminate(ctx) })

	dbURL, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}

	ai := &fakeAI{emb: make([]float32, 1536)}
	appInstance, err := app.New(&config.Config{
		DatabaseURL: dbURL,
		JWTSecret:   []byte("test"),
	}, ai)
	if err != nil {
		t.Fatalf("setup app: %v", err)
	}
	t.Cleanup(func() { appInstance.Close() })

	var srv *httptest.Server
	appInstance.Listen(func(_ uint16, router http.Handler) error {
		srv = httptest.NewServer(router)
		t.Cleanup(srv.Close)
		return nil
	})

	// prepare a default user for convenience
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	_, err = db.Exec(`INSERT INTO users(email, password_hash, created_at, updated_at) VALUES($1,$2,now(),now())`,
		"user@example.com", string(hashed))
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	return &testApp{srv: srv, ai: ai}
}

func TestAuthAPI(t *testing.T) {
	app := setupApp(t)
	regBody := strings.NewReader(`{"email":"u@example.com","password":"pw"}`)
	resp, err := http.Post(app.srv.URL+"/api/register", "application/json", regBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	loginBody := strings.NewReader(`{"email":"u@example.com","password":"pw"}`)
	resp, err = http.Post(app.srv.URL+"/api/login", "application/json", loginBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var data map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&data))
	assert.NotEmpty(t, data["token"])
}

type fakeAI struct {
	emb        []float32
	lastPrompt string
}

func (f *fakeAI) CreateEmbeddings(ctx context.Context, req go_openai.EmbeddingRequestConverter) (go_openai.EmbeddingResponse, error) {
	return go_openai.EmbeddingResponse{Data: []go_openai.Embedding{{Embedding: f.emb}}}, nil
}

func (f *fakeAI) CreateChatCompletion(ctx context.Context, req go_openai.ChatCompletionRequest) (go_openai.ChatCompletionResponse, error) {
	f.lastPrompt = req.Messages[len(req.Messages)-1].Content
	return go_openai.ChatCompletionResponse{Choices: []go_openai.ChatCompletionChoice{{Message: go_openai.ChatCompletionMessage{Content: "ok"}}}}, nil
}

func TestQnAAPI(t *testing.T) {
	app := setupApp(t)

	// create KB
	resp, err := http.Post(app.srv.URL+"/api/kbs", "application/json", strings.NewReader(`{"name":"demo"}`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var kb handlers.KB
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&kb))

	// upload file
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "a.txt")
	fw.Write([]byte("hello world"))
	mw.Close()
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/kbs/%d/files", app.srv.URL, kb.ID), &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// ask question
	resp, err = http.Post(fmt.Sprintf("%s/api/kbs/%d/ask", app.srv.URL, kb.ID), "application/json", strings.NewReader(`{"question":"hi"}`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var ans map[string]any
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&ans))
	assert.Equal(t, "ok", ans["answer"])
}

func TestPDFUploadDownloadRoundtrip(t *testing.T) {
	app := setupApp(t)

	// create KB
	resp, err := http.Post(app.srv.URL+"/api/kbs", "application/json", strings.NewReader(`{"name":"demo"}`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var kb handlers.KB
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&kb))

	// read PDF file from testdata
	pdfPath := "internal/handlers/testdata/pdf_test.pdf"
	validPDF, err := os.ReadFile(pdfPath)
	assert.NoError(t, err)

	// upload PDF file
	var pdfBuf bytes.Buffer
	pdfMw := multipart.NewWriter(&pdfBuf)
	pdfFw, _ := pdfMw.CreateFormFile("file", "test.pdf")
	pdfFw.Write(validPDF)
	pdfMw.Close()
	pdfReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/kbs/%d/files", app.srv.URL, kb.ID), &pdfBuf)
	pdfReq.Header.Set("Content-Type", pdfMw.FormDataContentType())
	pdfResp, err := http.DefaultClient.Do(pdfReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, pdfResp.StatusCode)

	// roundtrip: download the PDF and compare bytes
	resp, err = http.Get(fmt.Sprintf("%s/api/kbs/%d/files", app.srv.URL, kb.ID))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var files []struct{ Name, Slug string }
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&files))
	var slug string
	for _, f := range files {
		if f.Name == "test.pdf" {
			slug = f.Slug
		}
	}
	assert.NotEmpty(t, slug)

	dlResp, err := http.Get(fmt.Sprintf("%s/api/kbs/%d/files/%s", app.srv.URL, kb.ID, slug))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, dlResp.StatusCode)
	dlBytes, err := io.ReadAll(dlResp.Body)
	assert.NoError(t, err)
	assert.Equal(t, validPDF, dlBytes)
}

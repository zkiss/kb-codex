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

// testUser represents a user with their authentication token
type testUser struct {
	Email string
	Token string
}

// testKB represents a knowledge base created by a user
type testKB struct {
	ID   int64
	Name string
	User *testUser
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

// createUserAndToken creates a new user and returns their authentication token
func (app *testApp) createUserAndToken(t *testing.T, email, password string) *testUser {
	t.Helper()

	// Register user
	regBody := strings.NewReader(fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password))
	resp, err := http.Post(app.srv.URL+"/api/register", "application/json", regBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Login to get token
	loginBody := strings.NewReader(fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password))
	resp, err = http.Post(app.srv.URL+"/api/login", "application/json", loginBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginData map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&loginData))
	token := loginData["token"]
	assert.NotEmpty(t, token)

	return &testUser{Email: email, Token: token}
}

// createKB creates a knowledge base for a user and returns it
func (app *testApp) createKB(t *testing.T, user *testUser, name string) *testKB {
	t.Helper()

	kbReq, err := http.NewRequest("POST", app.srv.URL+"/api/kbs", strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, name)))
	assert.NoError(t, err)
	kbReq.Header.Set("Content-Type", "application/json")
	kbReq.Header.Set("Authorization", "Bearer "+user.Token)

	resp, err := http.DefaultClient.Do(kbReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var kb handlers.KB
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&kb))

	return &testKB{ID: kb.ID, Name: name, User: user}
}

// askQuestion asks a question on a knowledge base and returns the response
func (app *testApp) askQuestion(t *testing.T, kb *testKB, question string) map[string]interface{} {
	t.Helper()

	questionReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/kbs/%d/ask", app.srv.URL, kb.ID), strings.NewReader(fmt.Sprintf(`{"question":"%s"}`, question)))
	assert.NoError(t, err)
	questionReq.Header.Set("Content-Type", "application/json")
	questionReq.Header.Set("Authorization", "Bearer "+kb.User.Token)

	resp, err := http.DefaultClient.Do(questionReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var answer map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&answer))
	return answer
}

// uploadFile uploads a file to a knowledge base
func (app *testApp) uploadFile(t *testing.T, kb *testKB, filename string, content []byte) {
	t.Helper()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", filename)
	fw.Write(content)
	mw.Close()

	uploadReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/kbs/%d/files", app.srv.URL, kb.ID), &buf)
	assert.NoError(t, err)
	uploadReq.Header.Set("Content-Type", mw.FormDataContentType())
	uploadReq.Header.Set("Authorization", "Bearer "+kb.User.Token)

	resp, err := http.DefaultClient.Do(uploadReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// makeRequest makes an HTTP request, optionally with authentication
func (app *testApp) makeRequest(t *testing.T, method, path string, user *testUser, body io.Reader) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, app.srv.URL+path, body)
	assert.NoError(t, err)

	// Add authentication if user is provided
	if user != nil {
		req.Header.Set("Authorization", "Bearer "+user.Token)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	return resp
}

func TestAuthAPI(t *testing.T) {
	app := setupApp(t)
	user := app.createUserAndToken(t, "u@example.com", "pw")
	assert.NotEmpty(t, user.Token)
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

	// Create user and knowledge base
	user := app.createUserAndToken(t, "test@example.com", "password")
	kb := app.createKB(t, user, "demo")

	// Upload file
	app.uploadFile(t, kb, "a.txt", []byte("hello world"))

	// Ask question
	answer := app.askQuestion(t, kb, "hi")
	assert.Equal(t, "ok", answer["answer"])
}

func TestPDFUploadDownloadRoundtrip(t *testing.T) {
	app := setupApp(t)

	// Create user and knowledge base
	user := app.createUserAndToken(t, "pdf@example.com", "password")
	kb := app.createKB(t, user, "demo")

	// Read PDF file from testdata
	pdfPath := "internal/handlers/testdata/pdf_test.pdf"
	validPDF, err := os.ReadFile(pdfPath)
	assert.NoError(t, err)

	// Upload PDF file
	app.uploadFile(t, kb, "test.pdf", validPDF)

	// Download the PDF and compare bytes
	resp := app.makeRequest(t, "GET", fmt.Sprintf("/api/kbs/%d/files", kb.ID), user, nil)
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

	// Download the actual file
	dlResp := app.makeRequest(t, "GET", fmt.Sprintf("/api/kbs/%d/files/%s", kb.ID, slug), user, nil)
	assert.Equal(t, http.StatusOK, dlResp.StatusCode)
	dlBytes, err := io.ReadAll(dlResp.Body)
	assert.NoError(t, err)
	assert.Equal(t, validPDF, dlBytes)
}

func TestUnauthenticatedAccess(t *testing.T) {
	app := setupApp(t)

	// Create user and knowledge base
	user := app.createUserAndToken(t, "auth@example.com", "password")
	kb := app.createKB(t, user, "test-kb")

	// Test unauthenticated access to various endpoints
	resp := app.makeRequest(t, "GET", "/api/kbs", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "list KBs without auth")

	resp = app.makeRequest(t, "POST", "/api/kbs", nil, strings.NewReader(`{"name":"unauthorized"}`))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "create KB without auth")

	resp = app.makeRequest(t, "GET", fmt.Sprintf("/api/kbs/%d/files", kb.ID), nil, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "list files without auth")

	resp = app.makeRequest(t, "POST", fmt.Sprintf("/api/kbs/%d/ask", kb.ID), nil, strings.NewReader(`{"question":"test"}`))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "ask question without auth")
}

func TestCrossUserAccess(t *testing.T) {
	app := setupApp(t)

	// Create first user and KB
	user1 := app.createUserAndToken(t, "user1@example.com", "password")
	kb1 := app.createKB(t, user1, "user1-kb")

	// Create second user
	user2 := app.createUserAndToken(t, "user2@example.com", "password")

	// Test that user2 cannot access user1's KB
	resp := app.makeRequest(t, "GET", fmt.Sprintf("/api/kbs/%d/files", kb1.ID), user2, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "user2 cannot list user1's files")

	resp2 := app.makeRequest(t, "POST", fmt.Sprintf("/api/kbs/%d/ask", kb1.ID), user2, strings.NewReader(`{"question":"test"}`))
	assert.Equal(t, http.StatusForbidden, resp2.StatusCode, "user2 cannot ask questions on user1's KB")

	// Verify that user2 can access their own KB (should work)
	kb2 := app.createKB(t, user2, "user2-kb")
	resp3 := app.makeRequest(t, "GET", fmt.Sprintf("/api/kbs/%d/files", kb2.ID), user2, nil)
	assert.Equal(t, http.StatusOK, resp3.StatusCode)
}

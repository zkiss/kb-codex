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

	// Register and login to get a token
	regBody := strings.NewReader(`{"email":"test@example.com","password":"password"}`)
	resp, err := http.Post(app.srv.URL+"/api/register", "application/json", regBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	loginBody := strings.NewReader(`{"email":"test@example.com","password":"password"}`)
	resp, err = http.Post(app.srv.URL+"/api/login", "application/json", loginBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var loginData map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&loginData))
	token := loginData["token"]
	assert.NotEmpty(t, token)

	// create KB with authentication
	kbReq, err := http.NewRequest("POST", app.srv.URL+"/api/kbs", strings.NewReader(`{"name":"demo"}`))
	assert.NoError(t, err)
	kbReq.Header.Set("Content-Type", "application/json")
	kbReq.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(kbReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var kb handlers.KB
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&kb))

	// upload file with authentication
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "a.txt")
	fw.Write([]byte("hello world"))
	mw.Close()
	uploadReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/kbs/%d/files", app.srv.URL, kb.ID), &buf)
	uploadReq.Header.Set("Content-Type", mw.FormDataContentType())
	uploadReq.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(uploadReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// ask question with authentication
	questionReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/kbs/%d/ask", app.srv.URL, kb.ID), strings.NewReader(`{"question":"hi"}`))
	questionReq.Header.Set("Content-Type", "application/json")
	questionReq.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(questionReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var ans map[string]any
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&ans))
	assert.Equal(t, "ok", ans["answer"])
}

func TestPDFUploadDownloadRoundtrip(t *testing.T) {
	app := setupApp(t)

	// Register and login to get a token
	regBody := strings.NewReader(`{"email":"pdf@example.com","password":"password"}`)
	resp, err := http.Post(app.srv.URL+"/api/register", "application/json", regBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	loginBody := strings.NewReader(`{"email":"pdf@example.com","password":"password"}`)
	resp, err = http.Post(app.srv.URL+"/api/login", "application/json", loginBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var loginData map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&loginData))
	token := loginData["token"]
	assert.NotEmpty(t, token)

	// create KB with authentication
	kbReq, err := http.NewRequest("POST", app.srv.URL+"/api/kbs", strings.NewReader(`{"name":"demo"}`))
	assert.NoError(t, err)
	kbReq.Header.Set("Content-Type", "application/json")
	kbReq.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(kbReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var kb handlers.KB
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&kb))

	// read PDF file from testdata
	pdfPath := "internal/handlers/testdata/pdf_test.pdf"
	validPDF, err := os.ReadFile(pdfPath)
	assert.NoError(t, err)

	// upload PDF file with authentication
	var pdfBuf bytes.Buffer
	pdfMw := multipart.NewWriter(&pdfBuf)
	pdfFw, _ := pdfMw.CreateFormFile("file", "test.pdf")
	pdfFw.Write(validPDF)
	pdfMw.Close()
	pdfReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/kbs/%d/files", app.srv.URL, kb.ID), &pdfBuf)
	pdfReq.Header.Set("Content-Type", pdfMw.FormDataContentType())
	pdfReq.Header.Set("Authorization", "Bearer "+token)
	pdfResp, err := http.DefaultClient.Do(pdfReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, pdfResp.StatusCode)

	// roundtrip: download the PDF and compare bytes with authentication
	downloadReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/kbs/%d/files", app.srv.URL, kb.ID), nil)
	downloadReq.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(downloadReq)
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

	dlReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/kbs/%d/files/%s", app.srv.URL, kb.ID, slug), nil)
	dlReq.Header.Set("Authorization", "Bearer "+token)
	dlResp, err := http.DefaultClient.Do(dlReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, dlResp.StatusCode)
	dlBytes, err := io.ReadAll(dlResp.Body)
	assert.NoError(t, err)
	assert.Equal(t, validPDF, dlBytes)
}

func TestUnauthenticatedAccess(t *testing.T) {
	app := setupApp(t)

	// Register and login to get a token
	regBody := strings.NewReader(`{"email":"auth@example.com","password":"password"}`)
	resp, err := http.Post(app.srv.URL+"/api/register", "application/json", regBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	loginBody := strings.NewReader(`{"email":"auth@example.com","password":"password"}`)
	resp, err = http.Post(app.srv.URL+"/api/login", "application/json", loginBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var loginData map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&loginData))
	token := loginData["token"]
	assert.NotEmpty(t, token)

	// create KB with authentication
	kbReq, err := http.NewRequest("POST", app.srv.URL+"/api/kbs", strings.NewReader(`{"name":"test-kb"}`))
	assert.NoError(t, err)
	kbReq.Header.Set("Content-Type", "application/json")
	kbReq.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(kbReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var kb handlers.KB
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&kb))

	// Test unauthenticated access to various endpoints
	testCases := []struct {
		name     string
		method   string
		url      string
		body     io.Reader
		expected int
	}{
		{
			name:     "list KBs without auth",
			method:   "GET",
			url:      app.srv.URL + "/api/kbs",
			expected: http.StatusUnauthorized,
		},
		{
			name:     "create KB without auth",
			method:   "POST",
			url:      app.srv.URL + "/api/kbs",
			body:     strings.NewReader(`{"name":"unauthorized"}`),
			expected: http.StatusUnauthorized,
		},
		{
			name:     "list files without auth",
			method:   "GET",
			url:      fmt.Sprintf("%s/api/kbs/%d/files", app.srv.URL, kb.ID),
			expected: http.StatusUnauthorized,
		},
		{
			name:     "ask question without auth",
			method:   "POST",
			url:      fmt.Sprintf("%s/api/kbs/%d/ask", app.srv.URL, kb.ID),
			body:     strings.NewReader(`{"question":"test"}`),
			expected: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, tc.body)
			assert.NoError(t, err)
			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, resp.StatusCode, "Expected %d for %s", tc.expected, tc.name)
		})
	}
}

func TestCrossUserAccess(t *testing.T) {
	app := setupApp(t)

	// Create first user and KB
	regBody1 := strings.NewReader(`{"email":"user1@example.com","password":"password"}`)
	resp, err := http.Post(app.srv.URL+"/api/register", "application/json", regBody1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	loginBody1 := strings.NewReader(`{"email":"user1@example.com","password":"password"}`)
	resp, err = http.Post(app.srv.URL+"/api/login", "application/json", loginBody1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var loginData1 map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&loginData1))
	token1 := loginData1["token"]
	assert.NotEmpty(t, token1)

	// Create KB for user1
	kbReq1, err := http.NewRequest("POST", app.srv.URL+"/api/kbs", strings.NewReader(`{"name":"user1-kb"}`))
	assert.NoError(t, err)
	kbReq1.Header.Set("Content-Type", "application/json")
	kbReq1.Header.Set("Authorization", "Bearer "+token1)
	resp, err = http.DefaultClient.Do(kbReq1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var kb1 handlers.KB
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&kb1))

	// Create second user
	regBody2 := strings.NewReader(`{"email":"user2@example.com","password":"password"}`)
	resp, err = http.Post(app.srv.URL+"/api/register", "application/json", regBody2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	loginBody2 := strings.NewReader(`{"email":"user2@example.com","password":"password"}`)
	resp, err = http.Post(app.srv.URL+"/api/login", "application/json", loginBody2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var loginData2 map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&loginData2))
	token2 := loginData2["token"]
	assert.NotEmpty(t, token2)

	// Test that user2 cannot access user1's KB
	testCases := []struct {
		name     string
		method   string
		url      string
		body     io.Reader
		expected int
	}{
		{
			name:     "user2 cannot list user1's files",
			method:   "GET",
			url:      fmt.Sprintf("%s/api/kbs/%d/files", app.srv.URL, kb1.ID),
			expected: http.StatusForbidden,
		},
		{
			name:     "user2 cannot ask questions on user1's KB",
			method:   "POST",
			url:      fmt.Sprintf("%s/api/kbs/%d/ask", app.srv.URL, kb1.ID),
			body:     strings.NewReader(`{"question":"test"}`),
			expected: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, tc.body)
			assert.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+token2)
			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, resp.StatusCode, "Expected %d for %s", tc.expected, tc.name)
		})
	}

	// Verify that user2 can access their own KB (should work)
	kbReq2, err := http.NewRequest("POST", app.srv.URL+"/api/kbs", strings.NewReader(`{"name":"user2-kb"}`))
	assert.NoError(t, err)
	kbReq2.Header.Set("Content-Type", "application/json")
	kbReq2.Header.Set("Authorization", "Bearer "+token2)
	resp, err = http.DefaultClient.Do(kbReq2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var kb2 handlers.KB
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&kb2))

	// Verify user2 can access their own KB
	listReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/kbs/%d/files", app.srv.URL, kb2.ID), nil)
	assert.NoError(t, err)
	listReq.Header.Set("Authorization", "Bearer "+token2)
	resp, err = http.DefaultClient.Do(listReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	jwtSecret := []byte("test-secret")

	// Create a test token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": float64(123),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedUserID int64
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusOK,
			expectedUserID: 123,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: 0,
		},
		{
			name:           "invalid bearer format",
			authHeader:     "Invalid " + tokenString,
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: 0,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedUserID int64
			var capturedContext context.Context

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				userID, ok := GetUserID(r.Context())
				if ok {
					capturedUserID = userID
				}
				w.WriteHeader(http.StatusOK)
			})

			middleware := AuthMiddleware(jwtSecret)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedUserID, capturedUserID)
				assert.NotNil(t, capturedContext)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, int64(123))

	userID, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, int64(123), userID)

	// Test with context without user ID
	emptyCtx := context.Background()
	userID, ok = GetUserID(emptyCtx)
	assert.False(t, ok)
	assert.Equal(t, int64(0), userID)
}

func TestRequireUserID(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, int64(123))

	userID, err := RequireUserID(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), userID)

	// Test with context without user ID
	emptyCtx := context.Background()
	userID, err = RequireUserID(emptyCtx)
	assert.Error(t, err)
	assert.Equal(t, int64(0), userID)
	assert.Contains(t, err.Error(), "user ID not found in context")
}

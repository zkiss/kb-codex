package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "time"

    "github.com/golang-jwt/jwt/v4"
    "golang.org/x/crypto/bcrypt"

//     "github.com/zkiss/kb-codex/internal/models"
)

// AuthHandler handles user authentication actions.
type AuthHandler struct {
    DB        *sql.DB
    JWTSecret []byte
}

// NewAuthHandler returns a new AuthHandler.
func NewAuthHandler(db *sql.DB, jwtSecret []byte) *AuthHandler {
    return &AuthHandler{
        DB:        db,
        JWTSecret: jwtSecret,
    }
}

// Credentials represents user login or registration payload.
type Credentials struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// Register registers a new user with email and password.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var creds Credentials
    if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
        http.Error(w, "invalid request payload", http.StatusBadRequest)
        return
    }
    if creds.Email == "" || creds.Password == "" {
        http.Error(w, "email and password required", http.StatusBadRequest)
        return
    }
    hashed, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "could not hash password", http.StatusInternalServerError)
        return
    }
    now := time.Now().UTC()
    _, err = h.DB.Exec(
        `INSERT INTO users(email, password_hash, created_at, updated_at)
         VALUES ($1, $2, $3, $4)`,
        creds.Email, string(hashed), now, now,
    )
    if err != nil {
        http.Error(w, "could not create user: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
}

// Login authenticates a user and returns a JWT token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var creds Credentials
    if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
        http.Error(w, "invalid request payload", http.StatusBadRequest)
        return
    }
    if creds.Email == "" || creds.Password == "" {
        http.Error(w, "email and password required", http.StatusBadRequest)
        return
    }
    var id int64
    var storedHash string
    err := h.DB.QueryRow(`SELECT id, password_hash FROM users WHERE email=$1`, creds.Email).Scan(&id, &storedHash)
    if err == sql.ErrNoRows {
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }
    if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(creds.Password)); err != nil {
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub": id,
        "exp": time.Now().Add(24 * time.Hour).Unix(),
    })
    tokenString, err := token.SignedString(h.JWTSecret)
    if err != nil {
        http.Error(w, "could not generate token", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

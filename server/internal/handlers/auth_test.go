package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))

	h := NewAuthHandler(db, []byte("secret"))

	body := `{"email":"a@b.com","password":"pw"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock: %v", err)
	}
	defer db.Close()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	rows := sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(1, string(hashed))
	mock.ExpectQuery("SELECT id, password_hash FROM users").WillReturnRows(rows)

	h := NewAuthHandler(db, []byte("secret"))
	body := `{"email":"a@b.com","password":"pw"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotEmpty(t, resp["token"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

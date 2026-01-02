package main

import (
	"bytes"

	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// --- Helper for DB Mocks ---
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, Handler) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	h := Handler{DB: db}
	return db, mock, h
}

// --- HANDLER TESTS ---

func TestHealth(t *testing.T) {
	_, _, h := setupMockDB(t)
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	h.Health(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateUser_Success(t *testing.T) {
	db, mock, h := setupMockDB(t)
	defer db.Close()

	newUser := User{ID: "u1", Username: "Test", Email: "t@e.com"}
	body, _ := json.Marshal(newUser)

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users`)).
		WithArgs(newUser.ID, newUser.Username, newUser.Email).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	_, _, h := setupMockDB(t)

	// Malformed JSON
	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(`{ "id": `))
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateUser_DBError(t *testing.T) {
	db, mock, h := setupMockDB(t)
	defer db.Close()

	newUser := User{ID: "u1", Username: "Test", Email: "t@e.com"}
	body, _ := json.Marshal(newUser)

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users`)).
		WillReturnError(errors.New("db down"))

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetUser_Success(t *testing.T) {
	db, mock, h := setupMockDB(t)
	defer db.Close()

	u := User{ID: "u1", Username: "Test", Email: "t@e.com", CreatedAt: time.Now()}
	rows := sqlmock.NewRows([]string{"id", "username", "email", "created_at"}).
		AddRow(u.ID, u.Username, u.Email, u.CreatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, email, created_at FROM users`)).
		WithArgs("u1").
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/users/u1", nil)
	rec := httptest.NewRecorder()
	
	// Need router for mux.Vars
	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", h.GetUser)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetUser_NotFound(t *testing.T) {
	db, mock, h := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs("unknown").
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/users/unknown", nil)
	rec := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", h.GetUser)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetUser_DBError(t *testing.T) {
	db, mock, h := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs("u1").
		WillReturnError(errors.New("connection failed"))

	req := httptest.NewRequest("GET", "/users/u1", nil)
	rec := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", h.GetUser)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// --- AUTH MIDDLEWARE TESTS ---

func TestExtractUser_MissingHeader(t *testing.T) {
	// Dummy handler to verify if we get through
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	ExtractUser(nextHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestExtractUser_InvalidToken(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	rec := httptest.NewRecorder()

	ExtractUser(nextHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestExtractUser_Success(t *testing.T) {
	// Create a dummy unsigned token
	claims := Claims{
		Sub:               "user-123",
		Email:             "test@test.com",
		PreferredUsername: "tester",
		RegisteredClaims:  jwt.RegisteredClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// We use "secret" but the parser is ParseUnverified so it won't actually check signature
	tokenStr, _ := token.SignedString([]byte("secret"))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Context was set
		assert.Equal(t, "user-123", r.Context().Value("userID"))
		assert.Equal(t, "test@test.com", r.Context().Value("email"))
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	ExtractUser(nextHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- METRICS MIDDLEWARE TEST ---

func TestMetricsMiddleware(t *testing.T) {
	// Only testing that it doesn't panic and passes request through
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	req := httptest.NewRequest("GET", "/some-route", nil)
	rec := httptest.NewRecorder()

	metricsMiddleware(nextHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
}
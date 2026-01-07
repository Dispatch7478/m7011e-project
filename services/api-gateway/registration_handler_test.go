package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationHandler_Handle(t *testing.T) {
	// mock us
	userService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/register", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
	}))
	defer userService.Close()

	// mock keycloak
	keycloakServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// mock admin token
		if strings.Contains(r.URL.Path, "/openid-connect/token") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock-admin-token",
				"expires_in":   300,
			})
			return
		}

		// mock creating a user
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/users") {
			// Return path to new user (gocloak behaviour) or just 201
			w.Header().Set("Location", "http://keycloak/user/new-user-id-123")
			w.WriteHeader(http.StatusCreated)
			return
		}

		// mock reset password
		if strings.Contains(r.URL.Path, "reset-password") {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer keycloakServer.Close()

	// setup handler (mock server url )
	// Note: We use the mock server URL for Keycloak
	kcClient := NewKeycloakClient(keycloakServer.URL, "test-realm")
	handler := &RegistrationHandler{
		Keycloak:    kcClient,
		UserService: userService.URL,
	}

	// send rq
	e := echo.New()
	reqBody := `{"username":"john", "email":"john@test.com", "password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestRegistrationHandler_InvalidBody(t *testing.T) {
	handler := &RegistrationHandler{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString("{invalid-json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)

	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}
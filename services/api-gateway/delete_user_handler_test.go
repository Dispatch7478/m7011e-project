package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDeletionHandler_Handle_Success(t *testing.T) {
	// 1. Mock User Service
	// Expects a DELETE request at /users/{id}
	userService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/test-user-123", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer userService.Close()

	// 2. Mock Keycloak
	keycloakServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Mock Admin Login (POST /realms/master/protocol/openid-connect/token)
		if strings.Contains(r.URL.Path, "/openid-connect/token") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock-admin-token",
				"expires_in":   300,
			})
			return
		}

		// Mock Delete User (DELETE /admin/realms/{realm}/users/{id})
		if r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/users/") {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer keycloakServer.Close()

	// 3. Setup Handler
	kcClient := NewKeycloakClient(keycloakServer.URL, "test-realm")
	handler := &DeletionHandler{
		Keycloak:    kcClient,
		UserService: userService.URL,
	}

	// 4. Create Request
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me", nil)
	// Simulate the header injected by AuthMiddleware
	req.Header.Set("X-User-Id", "test-user-123")
	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 5. Execute
	err := handler.Handle(c)

	// 6. Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeletionHandler_MissingUserID(t *testing.T) {
	handler := &DeletionHandler{}
	e := echo.New()
	
	// Request without X-User-Id header
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)

	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}

func TestDeletionHandler_KeycloakFailure(t *testing.T) {
	// Mock Keycloak to fail on login (or delete)
	keycloakServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer keycloakServer.Close()

	kcClient := NewKeycloakClient(keycloakServer.URL, "test-realm")
	handler := &DeletionHandler{
		Keycloak: kcClient,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me", nil)
	req.Header.Set("X-User-Id", "test-user-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)

	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	// Should fail because it can't login as admin
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}

func TestDeletionHandler_UserServiceFailure(t *testing.T) {
	// 1. Mock User Service to fail
	userService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer userService.Close()

	// 2. Mock Keycloak (Successful login and delete)
	keycloakServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/openid-connect/token") {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "mock", "expires_in": 300})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer keycloakServer.Close()

	// 3. Setup
	kcClient := NewKeycloakClient(keycloakServer.URL, "test-realm")
	handler := &DeletionHandler{
		Keycloak:    kcClient,
		UserService: userService.URL,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me", nil)
	req.Header.Set("X-User-Id", "test-user-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 4. Execute
	err := handler.Handle(c)

	// 5. Assert
	// If User Service fails (500), the Gateway should return 500 (or the specific error mapped in handler)
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}
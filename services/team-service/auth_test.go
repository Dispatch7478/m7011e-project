package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestExtractUser_UnauthorizedWhenMissingHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("next handler should not be called")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	ExtractUser(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestExtractUser_UnauthorizedWhenBadToken(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("next handler should not be called")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer definitely-not-a-jwt")

	ExtractUser(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestExtractUser_SetsContextAndCallsNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		uid, ok := userIDFromCtx(r.Context())
		if !ok || uid != "user-123" {
			t.Fatalf("expected user-123 in ctx, got %q ok=%v", uid, ok)
		}

		// Also verify email/username context values exist
		if v := r.Context().Value(ctxEmail); v != "u@example.com" {
			t.Fatalf("expected ctxEmail u@example.com, got %#v", v)
		}
		if v := r.Context().Value(ctxUsername); v != "demo" {
			t.Fatalf("expected ctxUsername demo, got %#v", v)
		}

		w.WriteHeader(http.StatusOK)
	})

	claims := &Claims{
		Sub:               "user-123",
		Email:             "u@example.com",
		PreferredUsername: "demo",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("secret-does-not-matter"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	ExtractUser(next).ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected next handler to be called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestUserIDFromCtx_FalseWhenMissing(t *testing.T) {
	uid, ok := userIDFromCtx(context.Background())
	if ok || uid != "" {
		t.Fatalf("expected ok=false and empty uid, got %q ok=%v", uid, ok)
	}
}

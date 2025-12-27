package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	ctxUserID   ctxKey = "userID"
	ctxEmail    ctxKey = "email"
	ctxUsername ctxKey = "username"
)

type Claims struct {
	Sub               string `json:"sub"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	jwt.RegisteredClaims
}

// DEV MODE: does NOT verify signature. Swap later.
func ExtractUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, _, err := jwt.NewParser().ParseUnverified(tokenStr, &Claims{})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || claims.Sub == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, claims.Sub)
		ctx = context.WithValue(ctx, ctxEmail, claims.Email)
		ctx = context.WithValue(ctx, ctxUsername, claims.PreferredUsername)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userIDFromCtx(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxUserID)
	s, ok := v.(string)
	return s, ok && s != ""
}

package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Sub               string `json:"sub"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	jwt.RegisteredClaims
}

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
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// attach to request context
		ctx := context.WithValue(r.Context(), "userID", claims.Sub)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "username", claims.PreferredUsername)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

package main

import (
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(provider *oidc.Provider) echo.MiddlewareFunc {
	verifier := provider.Verifier(&oidc.Config{
		SkipClientIDCheck: true, // We trust the issuer (Keycloak), any client is fine
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract Token
			rawToken := c.Request().Header.Get("Authorization")
			if rawToken == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
			}

			// Remove "Bearer " prefix if present
			parts := strings.Split(rawToken, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token format")
			}

			// Verify Token
			idToken, err := verifier.Verify(c.Request().Context(), parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token: "+err.Error())
			}

			// Extract User ID (Subject)
			// This is the unique ID from Keycloak (e.g., a UUID)
			userID := idToken.Subject

			// Inject Header for downstream service
			// The User Service will trust this header
			c.Request().Header.Set("X-User-Id", userID)

			return next(c)
		}
	}
}

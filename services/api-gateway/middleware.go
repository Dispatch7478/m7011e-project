package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
)

// User represents the structure of a user object from the user-service
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// fetchUserDetails makes a request to the user-service to get user details
func fetchUserDetails(ctx context.Context, userServiceURL, userID string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/users/%s", userServiceURL, userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user-service request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call user-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user-service returned non-ok status: %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user-service response: %w", err)
	}

	return &user, nil
}

func AuthMiddleware(provider *oidc.Provider, userServiceURL string) echo.MiddlewareFunc {
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

			user, err := fetchUserDetails(c.Request().Context(), userServiceURL, userID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch user details: "+err.Error())
			}

			// Inject Headers for downstream service
			c.Request().Header.Set("X-User-Id", userID)
			c.Request().Header.Set("X-User-Name", user.Username)

			return next(c)
		}
	}
}

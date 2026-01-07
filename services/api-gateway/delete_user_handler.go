package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type DeletionHandler struct {
	Keycloak    *KeycloakClient
	UserService string
}

func (h *DeletionHandler) Handle(c echo.Context) error {
	// 1. Get User ID from context (extracted by AuthMiddleware)
	userID := c.Request().Header.Get("X-User-Id")
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	ctx := c.Request().Context()

	// 2. Delete from Keycloak
	// We need an admin token to perform deletion
	adminUser := getEnv("KEYCLOAK_ADMIN_USER", "admin")
	adminPassword := getEnv("KEYCLOAK_ADMIN_PASSWORD", "admin")

	token, err := h.Keycloak.Client.LoginAdmin(ctx, adminUser, adminPassword, "master")
	if err != nil {
		c.Logger().Errorf("Failed to login as admin: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	err = h.Keycloak.Client.DeleteUser(ctx, token.AccessToken, h.Keycloak.Realm, userID)
	if err != nil {
		c.Logger().Errorf("Failed to delete user from Keycloak: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete user account")
	}

	// 3. Delete from User Service
	// We act as a proxy here, calling the user-service's delete endpoint
	req, err := http.NewRequestWithContext(ctx, "DELETE", h.UserService+"/users/"+userID, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create request to user-service")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Logger().Errorf("Failed to call user-service delete: %v", err)
		// Note: Keycloak deletion already succeeded, so we have a partial consistency state.
		// ideally, we would log this for manual cleanup, but for now we report error.
		return echo.NewHTTPError(http.StatusBadGateway, "Account deleted from login, but failed to clean up profile data")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		c.Logger().Errorf("User-service returned unexpected status: %d", resp.StatusCode)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to clean up user profile")
	}

	return c.NoContent(http.StatusNoContent)
}
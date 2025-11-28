package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/labstack/echo/v4"
)

type RegistrationRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegistrationHandler struct {
	Keycloak    *KeycloakClient
	UserService string
}

func (h *RegistrationHandler) Handle(c echo.Context) error {
	var req RegistrationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Create the user in Keycloak
	user := gocloak.User{
		Username: gocloak.StringP(req.Username),
		Email:    gocloak.StringP(req.Email),
	}
	userID, err := h.Keycloak.CreateUser(c.Request().Context(), user, req.Password)
	if err != nil {
		c.Logger().Errorf("Failed to create user in Keycloak: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user in Keycloak")
	}

	// Register the user in the user-service
	userServiceReq := struct {
		ID       string `json:"id"` // From Keycloak.
		Username string `json:"username"`
		Email    string `json:"email"`
	}{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
	}

	body, err := json.Marshal(userServiceReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create request for user-service")
	}

	resp, err := http.Post(h.UserService+"/register", "application/json", bytes.NewBuffer(body))
	if err != nil || (resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK) {
		// Log the detailed error
		if err != nil {
			c.Logger().Errorf("Error registering user with user-service: %v", err)
		} else {
			c.Logger().Errorf("User-service returned non-successful status: %d", resp.StatusCode)
		}

		// Attempt to delete the user from Keycloak if the user-service registration fails
		_ = h.Keycloak.Client.DeleteUser(c.Request().Context(), "", h.Keycloak.Realm, userID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user in user-service")
	}

	return c.JSON(http.StatusCreated, nil)
}

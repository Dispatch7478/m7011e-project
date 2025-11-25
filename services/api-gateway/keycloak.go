package main

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
)

type KeycloakClient struct {
	Client *gocloak.GoCloak
	Realm  string
}

func NewKeycloakClient(url, realm string) *KeycloakClient {
	client := gocloak.NewClient(url)
	return &KeycloakClient{
		Client: client,
		Realm:  realm,
	}
}

func (k *KeycloakClient) CreateUser(ctx context.Context, user gocloak.User, password string) (string, error) {
	// An admin account is required to create users.
	// For this example, we'll use the admin account credentials from environment variables.
	// In a real-world scenario, you would use a service account with the appropriate permissions.
	adminUser := getEnv("KEYCLOAK_ADMIN_USER", "admin")
	adminPassword := getEnv("KEYCLOAK_ADMIN_PASSWORD", "admin")

	token, err := k.Client.LoginAdmin(ctx, adminUser, adminPassword, "master")
	if err != nil {
		return "", err
	}

	user.Enabled = gocloak.BoolP(true)
	userID, err := k.Client.CreateUser(ctx, token.AccessToken, k.Realm, user)
	if err != nil {
		return "", err
	}

	err = k.Client.SetPassword(ctx, token.AccessToken, userID, k.Realm, password, false)
	if err != nil {
		// Attempt to delete the user if setting the password fails
		_ = k.Client.DeleteUser(ctx, token.AccessToken, k.Realm, userID)
		return "", err
	}

	return userID, nil
}

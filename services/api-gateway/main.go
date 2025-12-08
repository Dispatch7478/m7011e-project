package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
)

func main() {
	// Load configuration
	config, err := LoadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	// Configuration via environment variables
	keycloakURL := getEnv("KEYCLOAK_URL", "")
	keycloakRealm := getEnv("KEYCLOAK_REALM", "t-hub")
	userServiceURL := getEnv("USER_SERVICE_URL", "")
	port := getEnv("PORT", ":8080")

	// Create a custom HTTP client that skips TLS verification
	// WARNING: This is insecure and should not be used in production!
	// It is acceptable for a development environment with staging certificates.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	ctx := oidc.ClientContext(context.Background(), client)

	// Initialize the OIDC provider
	provider, err := oidc.NewProvider(ctx, keycloakURL+"/realms/"+keycloakRealm)
	if err != nil {
		panic(err)
	}

	// Create the Keycloak client
	keycloakClient := NewKeycloakClient(keycloakURL, keycloakRealm)

	// Create the registration handler
	registrationHandler := &RegistrationHandler{
		Keycloak:    keycloakClient,
		UserService: userServiceURL,
	}

	// Create the router
	e := NewRouter(config, provider, registrationHandler)

	// Start the server
	e.Logger.Fatal(e.Start(port))
}

// Helper to read env vars
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
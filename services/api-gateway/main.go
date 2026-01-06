package main

import (
	"context"
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
	port := getEnv("PORT", ":8080")

	// Find the user-service URL from the config, making it the single source of truth
	var userServiceURL string
	for _, service := range config.Services {
		if service.Name == "user-service" {
			userServiceURL = service.URL
			break
		}
	}
	if userServiceURL == "" {
		panic("user-service URL not found in config.yaml")
	}

	// With production certificates, we use the default HTTP client which performs TLS verification.
	ctx := context.Background()

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
	e := NewRouter(config, provider, registrationHandler, userServiceURL)

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
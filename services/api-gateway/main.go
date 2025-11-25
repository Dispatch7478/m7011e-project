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
	keycloakURL := getEnv("KEYCLOAK_URL", "https://keycloak.ltu-m7011e-4.se")
	keycloakRealm := getEnv("KEYCLOAK_REALM", "t-hub")
	userServiceURL := getEnv("USER_SERVICE_URL", "http://user-service:8080")
	port := getEnv("PORT", ":8080")

	// Initialize the OIDC provider
	provider, err := oidc.NewProvider(context.Background(), keycloakURL+"/realms/"+keycloakRealm)
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
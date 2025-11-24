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
	keycloakURL := getEnv("KEYCLOAK_URL", "https://keycloak.ltu-m7011e-4.se/realms/t-hub")
	port := getEnv("PORT", ":8080")

	// Initialize the OIDC provider
	provider, err := oidc.NewProvider(context.Background(), keycloakURL)
	if err != nil {
		panic(err)
	}

	// Create the router
	e := NewRouter(config, provider)

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
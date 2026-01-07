package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-jose/go-jose.v2"
	"gopkg.in/go-jose/go-jose.v2/jwt"
)

func setupMockOIDCProvider(t *testing.T) (*oidc.Provider, string, jose.Signer) {
	// Generate RSA Key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create JWK
	// Use Public Key for verification
	jwk := jose.JSONWebKey{
		Key:       &key.PublicKey,
		KeyID:     "test-key",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	// Create Signer
	// Use Private Key for signing
	opts := (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "test-key")
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, opts)
	require.NoError(t, err)

	// Mock Server
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	// Serve Discovery Doc
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		// need to specify content-type otherwise oidc refuses to parse apparently
		w.Header().Set("Content-Type", "application/json") 
		json.NewEncoder(w).Encode(map[string]interface{}{
			"issuer":                 server.URL,
			"jwks_uri":               server.URL + "/keys",
			"response_types_supported": []string{"id_token"},
			"subject_types_supported":  []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})

	// Serve Keys
	mux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		// same as above
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}})
	})

	// Initialize Provider
	provider, err := oidc.NewProvider(context.Background(), server.URL)
	require.NoError(t, err)

	return provider, server.URL, signer
}

func TestAuthMiddleware(t *testing.T) {
	provider, issuer, signer := setupMockOIDCProvider(t)

	// Define test cases
	tests := []struct {
		name           string
		authHeader     string
		generateToken  bool
		expectedStatus int
	}{
		{
			name:           "Valid Token",
			generateToken:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Format",
			authHeader:     "Bearer",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Token String",
			authHeader:     "Bearer invalid.token.string",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			if tc.generateToken {
				// Create a valid JWT
				claims := map[string]interface{}{
					"sub":                "user-123",
					"iss":                issuer,
					"aud":                "test-client",
					"exp":                time.Now().Add(time.Hour).Unix(),
					"preferred_username": "testuser",
				}
				builder := jwt.Signed(signer).Claims(claims)
				tokenString, err := builder.CompactSerialize()
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+tokenString)
			} else if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			c := e.NewContext(req, rec)

			// The Handler to run if middleware passes
			handler := func(c echo.Context) error {
				// Verify headers were injected
				assert.Equal(t, "user-123", c.Request().Header.Get("X-User-Id"))
				assert.Equal(t, "testuser", c.Request().Header.Get("X-User-Name"))
				return c.String(http.StatusOK, "success")
			}

			middleware := AuthMiddleware(provider, "http://user-service")
			err := middleware(handler)(c)

			if tc.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
			} else {
				// Echo returns an error struct, check it
				if assert.Error(t, err) {
					he, ok := err.(*echo.HTTPError)
					assert.True(t, ok)
					assert.Equal(t, tc.expectedStatus, he.Code)
				}
			}
		})
	}
}
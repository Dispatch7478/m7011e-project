package main

import (
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewRouter(config *Config, provider *oidc.Provider, registrationHandler *RegistrationHandler) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Health Check
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Gateway Healthy")
	})

	// Registration endpoint (public)
	e.POST("/api/register", registrationHandler.Handle)

	for _, service := range config.Services {
		target, err := url.Parse(service.URL)
		if err != nil {
			e.Logger.Fatal("Invalid service URL: ", err)
		}

		// Group for protected API routes
		api := e.Group(service.Proxy.Prefix)
		api.Use(AuthMiddleware(provider))

		// Proxy configuration
		proxyConfig := middleware.ProxyConfig{
			Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
				{URL: target},
			}),
			Rewrite: map[string]string{
				service.Proxy.Prefix + "/*": "/$1",
			},
		}

		// Proxy the requests
		api.Group(service.Proxy.Rewrite, middleware.ProxyWithConfig(proxyConfig))
	}

	return e
}

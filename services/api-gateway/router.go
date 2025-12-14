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

	// CORS Middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://t-hub.ltu-m7011e-4.se"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Health Check
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "demo")
	})

	// Registration endpoint (public)
	e.POST("/api/register", registrationHandler.Handle)

	for _, service := range config.Services {
		target, err := url.Parse(service.URL)
		if err != nil {
			e.Logger.Fatal("Invalid service URL: ", err)
		}

		apiGroup := e.Group(service.Proxy.Path)
		apiGroup.Use(AuthMiddleware(provider))

		proxyConfig := middleware.ProxyConfig{
			Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
				{URL: target},
			}),
			Rewrite: map[string]string{
				service.Proxy.Path + "/*": "/$1",
				service.Proxy.Path:       "/",
			},
		}

		apiGroup.Use(middleware.ProxyWithConfig(proxyConfig))
	}

	return e
}

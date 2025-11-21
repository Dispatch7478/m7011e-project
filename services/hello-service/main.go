package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)
	e.GET("/health", health)
	
	// Metrics endpoint for Prometheus (REQ13)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

// Handler
func hello(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Hello! The pipeline works perfectly.",
		"service": "hello-service",
		"version": "1.0.0",
	})
}

func health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
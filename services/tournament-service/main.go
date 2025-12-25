package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
)

func main() {
	// Connect to Dependencies
	// Database
	dbPool, err := ConnectDB()
	if err != nil {
		log.Fatalf("Could not connect to Database: %v", err)
	}
	defer dbPool.Close()

	// RabbitMQ
	rmq, err := Connect()
	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ: %v", err)
	}
	defer rmq.Conn.Close()
	defer rmq.Channel.Close()

	// Setup Echo
	e := echo.New()

	// Middleware (Logging, Recover)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/health", HealthCheckHandler)
	
	e.POST("/tournaments", CreateTournamentHandler(dbPool, rmq))
	e.GET("/tournaments", GetAllTournamentsHandler(dbPool))

	e.POST("/tournaments/:id/register", RegisterTournamentHandler(dbPool))
	// Updaters
	e.PATCH("/tournaments/:id/status", UpdateTournamentStatusHandler(dbPool, rmq))
	e.GET("/api/tournaments/:id", GetTournamentHandler(dbPool))
	e.PUT("/api/tournaments/:id", UpdateTournamentDetailsHandler(dbPool, rmq))


	// Start Server
	port := "8080"
	e.Logger.Fatal(e.Start(":" + port))
}

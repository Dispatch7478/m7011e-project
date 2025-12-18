package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
)

// --- Main Entrypoint ---

func main() {
	// A. Connect to Dependencies
	// 1. Database
	dbPool, err := ConnectDB()
	if err != nil {
		log.Fatalf("Could not connect to Database: %v", err)
	}
	defer dbPool.Close()

	// 2. RabbitMQ
	rmq, err := Connect()
	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ: %v", err)
	}
	defer rmq.Conn.Close()
	defer rmq.Channel.Close()

	// B. Setup Echo
	e := echo.New()

	// Middleware (Logging, Recover)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// C. Routes
	e.POST("/tournaments", CreateTournamentHandler(dbPool, rmq))
	e.GET("/tournaments", GetAllTournamentsHandler(dbPool))
	e.GET("/health", HealthCheckHandler)

	// D. Start Server
	port := "8080"
	e.Logger.Fatal(e.Start(":" + port))
}

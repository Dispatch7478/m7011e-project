package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// 1. Database
	dbPool, err := ConnectDB()
	if err != nil {
		log.Fatalf("DB Error: %v", err)
	}
	defer dbPool.Close()

	// 2. RabbitMQ
	rmq, err := ConnectRabbitMQ()
	if err != nil {
		log.Fatalf("RabbitMQ Error: %v", err)
	}
	defer rmq.Conn.Close()
	defer rmq.Channel.Close()

	// 3. Config (for service URLs)
    // You'll need to create a simple Config struct loader similar to tournament-service
	tournamentServiceURL := os.Getenv("TOURNAMENT_SERVICE_URL") 
    if tournamentServiceURL == "" {
        // Fallback for local dev or hardcoded if prefered for MVP
        tournamentServiceURL = "http://tournament-service.t-hub-dev.svc.cluster.local:8080"
    }

	// 4. Echo Setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 5. Routes
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Bracket Service Healthy")
	})

    // Handler Initialization (We will create this next)
    h := &BracketHandler{DB: dbPool, RMQ: rmq, TournamentServiceURL: tournamentServiceURL}
    e.POST("/brackets/generate", h.GenerateBracket)
    e.GET("/brackets/:tournamentId", h.GetBracket)
	e.POST("/brackets/matches/:match_id/result", h.UpdateMatchResult)

	port := ":8080"
	e.Logger.Fatal(e.Start(port))
}
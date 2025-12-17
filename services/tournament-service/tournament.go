package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// --- Data Models ---

type Tournament struct {
	ID              string    `json:"id"`
	OrganizerID     string    `json:"organizer_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Game            string    `json:"game"`
	Format          string    `json:"format"`
	ParticipantType string    `json:"participant_type"`
	StartDate       time.Time `json:"start_date"`
	Status          string    `json:"status"`
	MinParticipants int       `json:"min_participants"`
	MaxParticipants int       `json:"max_participants"`
	Public          bool      `json:"public"`
}

type Event struct {
	EventType string      `json:"event_type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// --- Handlers ---

// CreateTournamentHandler now accepts DB pool and RabbitMQ service
func CreateTournamentHandler(db *pgxpool.Pool, rmq *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var t Tournament

		// 1. Bind (Parse) JSON
		if err := c.Bind(&t); err != nil {
			log.Printf("Failed to bind tournament data: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON input"})
		}

		// 2. Get Organizer ID from Header
		organizerID := c.Request().Header.Get("X-User-Id")
		if organizerID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing X-User-Id header"})
		}
		t.OrganizerID = organizerID

		// Validating "Power of 2" for participants
		if t.MaxParticipants < 2 || t.MaxParticipants > 16 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Max participants must be between 2 and 16"})
		}
		// Can fix later to add byes etc.
		if t.MaxParticipants%2 != 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Max participants must be a multiple of 2"})
		}

		// 3. Set Server-Side Defaults
		t.ID = uuid.New().String()
		t.Status = "draft" // Default status
		t.Public = true    // Default to public

		// 4. Insert into PostgreSQL
		query := `
			INSERT INTO tournaments 
			(id, organizer_id, name, description, game, format, start_date, status, min_participants, max_participants, public)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err := db.Exec(context.Background(), query,
			t.ID, t.OrganizerID, t.Name, t.Description, t.Game,
			t.Format, t.StartDate, t.Status, t.MinParticipants, t.MaxParticipants, t.Public,
		)

		if err != nil {
			log.Printf("Database Insert Error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save tournament"})
		}
		
		log.Printf("PERSISTED: Tournament '%s' (ID: %s)", t.Name, t.ID)

		// 5. Publish Event to RabbitMQ
		// Event Name: TournamentCreated
		// Routing Key: events.tournament.created
		event := Event{
			EventType: "TournamentCreated",
			Payload:   t,
			Timestamp: time.Now(),
		}

		eventBytes, _ := json.Marshal(event)

		// Passing the routing key as the first argument
		err = rmq.Publish("events.tournament.created", string(eventBytes))
		if err != nil {
			log.Printf("ERROR: Failed to publish event: %v", err)
			// Decide if this is fatal. For now, we log it but still return success for the DB save.
		}
			// 6. Return Success
			return c.JSON(http.StatusCreated, t)
		}
}

		

func GetAllTournamentsHandler(db *pgxpool.Pool) echo.HandlerFunc {

	return func(c echo.Context) error {
		query :=`
			SELECT id, organizer_id, name, description, game, format, start_date, status, min_participants, max_participants, public
			FROM tournaments
			WHERE public = true
		`

		rows, err := db.Query(context.Background(), query)

		if err != nil {
			log.Printf("Database Query Error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch tournaments"})
		}
		defer rows.Close()

		var tournaments []Tournament

		for rows.Next() {
			var t Tournament

			err := rows.Scan(&t.ID, &t.OrganizerID, &t.Name, &t.Description, &t.Game, &t.Format, &t.StartDate, &t.Status, &t.MinParticipants, &t.MaxParticipants, &t.Public)
			if err != nil {
				log.Printf("Row Scan Error: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process tournaments"})
			}
			tournaments = append(tournaments, t)
		}
		return c.JSON(http.StatusOK, tournaments)
	}
}


func HealthCheckHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Tournament Service Healthy")
}		
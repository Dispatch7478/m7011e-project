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
	ID                  string    `json:"id"`
	OrganizerID         string    `json:"organizer_id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Game                string    `json:"game"`
	Format              string    `json:"format"`
	ParticipantType     string    `json:"participant_type"`
	StartDate           time.Time `json:"start_date"`
	Status              string    `json:"status"`
	MinParticipants     int       `json:"min_participants"`
	MaxParticipants     int       `json:"max_participants"`
	Public              bool      `json:"public"`
	CurrentParticipants int       `json:"current_participants"`
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
			(id, organizer_id, name, description, game, format, participant_type, start_date, status, min_participants, max_participants, public)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		_, err := db.Exec(context.Background(), query,
			t.ID, t.OrganizerID, t.Name, t.Description, t.Game,
			t.Format, t.ParticipantType, t.StartDate, t.Status, t.MinParticipants, t.MaxParticipants, t.Public,
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

// func RegisterTournamentHandler(db *pgxpool.Pool) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		tournamentID := c.Param("id")
// 		participantID := c.Request().Header.Get("X-User-Id")

// 		if participantID == "" {
// 			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing X-User-Id header"})
// 		}

// 		// 1. Check if tournament exists, is public, is open for registration, and is not full
// 		var t Tournament
// 		// Note: pgxpool uses $1, $2, etc. as placeholders for query parameters.
// 		query := `SELECT id, status, public, max_participants FROM tournaments WHERE id = $1`
// 		err := db.QueryRow(context.Background(), query, tournamentID).Scan(&t.ID, &t.Status, &t.Public, &t.MaxParticipants)

// 		if err != nil {
// 			if err.Error() == "no rows in result set" {
// 				return c.JSON(http.StatusNotFound, map[string]string{"error": "Tournament not found"})
// 			}
// 			log.Printf("Database Query Error: %v", err)
// 			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check tournament details"})
// 		}

// 		if !t.Public {
// 			return c.JSON(http.StatusForbidden, map[string]string{"error": "Cannot register for private tournaments this way"})
// 		}

// 		if t.Status != "registration" {
// 			return c.JSON(http.StatusForbidden, map[string]string{"error": "Tournament is not open for registration"})
// 		}

// 		// 2. Check if tournament is full
// 		var count int
// 		countQuery := `SELECT count(*) FROM registrations WHERE tournament_id = $1`
// 		err = db.QueryRow(context.Background(), countQuery, tournamentID).Scan(&count)
// 		if err != nil {
// 			log.Printf("Database Query Error: %v", err)
// 			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check tournament registration count"})
// 		}

// 		if count >= t.MaxParticipants {
// 			return c.JSON(http.StatusForbidden, map[string]string{"error": "Tournament is full"})
// 		}

// 		// 3. Insert into registrations table
// 		// Note: pgxpool uses $1, $2, etc. as placeholders for query parameters.
// 		insertQuery := `
// 			INSERT INTO registrations (tournament_id, participant_id, participant_name, status)
// 			VALUES ($1, $2, $3, 'pending')
// 		`
// 		_, err = db.Exec(context.Background(), insertQuery, tournamentID, participantID, participantName)
// 		if err != nil {
// 			// Handle duplicate entry error (e.g., user already registered)
// 			if err.Error() == "ERROR: duplicate key value violates unique constraint \"registrations_pkey\" (SQLSTATE 23505)" {
// 				return c.JSON(http.StatusConflict, map[string]string{"error": "Already registered for this tournament"})
// 			}
// 			log.Printf("Database Insert Error: %v", err)
// 			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to register for tournament"})
// 		}

// 		return c.JSON(http.StatusCreated, map[string]string{"message": "Successfully registered for tournament"})
// 	}
// }

type RegistrationRequest struct {
	TeamID string `json:"team_id"` // Optional: Only for team tournaments
	Name   string `json:"name"`    // Required: Display name (User or Team name)
}


func GetAllTournamentsHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := `
			SELECT 
				t.id, t.organizer_id, t.name, 
				COALESCE(t.description, '') as description,
				t.game, t.format, t.participant_type, t.start_date, 
				t.status, t.min_participants, t.max_participants, t.public,
				COUNT(r.participant_id) as current_participants
			FROM tournaments t
			LEFT JOIN registrations r ON t.id = r.tournament_id
			WHERE t.public = true
			GROUP BY t.id
		`

		rows, err := db.Query(context.Background(), query)

		if err != nil {
			log.Printf("Database Query Error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch tournaments"})
		}
		defer rows.Close()

		// Empty slice in case there are no tournaments.
		tournaments := make([]Tournament, 0)

		for rows.Next() {
			var t Tournament

			err := rows.Scan(
				&t.ID, &t.OrganizerID, &t.Name, &t.Description, &t.Game,
				&t.Format, &t.ParticipantType, &t.StartDate, &t.Status, &t.MinParticipants,
				&t.MaxParticipants, &t.Public, &t.CurrentParticipants)

			if err != nil {
				log.Printf("Row Scan Error: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process tournaments"})
			}
			tournaments = append(tournaments, t)
		}
		debugBytes, _ := json.Marshal(tournaments)
    	log.Printf("DEBUG: GetAllTournamentsHandler found %d records. Payload: %s", len(tournaments), string(debugBytes))
		return c.JSON(http.StatusOK, tournaments)
	}
}

func RegisterTournamentHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		tournamentID := c.Param("id")
		userID := c.Request().Header.Get("X-User-Id")
		userName := c.Request().Header.Get("X-User-Name")

		if userID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing authentication"})
		}

		// 1. Parse Request Body (We need the name and optional Team ID)
		var req RegistrationRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		
		

		// 2. Fetch Tournament Details (Now including participant_type)
		var t Tournament
		query := `SELECT id, status, public, max_participants, participant_type FROM tournaments WHERE id = $1`
		err := db.QueryRow(context.Background(), query, tournamentID).Scan(
			&t.ID, &t.Status, &t.Public, &t.MaxParticipants, &t.ParticipantType,
		)

		if err != nil {
			if err.Error() == "no rows in result set" {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Tournament not found"})
			}
			log.Printf("Database Query Error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check tournament details"})
		}

		// If Individual, prefer the secure username from the Gateway header
		if t.ParticipantType == "individual" && userName != "" {
			req.Name = userName
		}

		if req.Name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Participant name is required"})
		}

		// 3. Validation Checks
		if t.Status != "registration_open" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Tournament is not open for registration"})
		}

		// 4. Determine Participant ID based on Type
		var participantID string

		if t.ParticipantType == "team" {
			// if req.TeamID == "" {
			// 	return c.JSON(http.StatusBadRequest, map[string]string{"error": "This is a team tournament. Team ID is required."})
			// }
			// participantID = req.TeamID
			return c.JSON(http.StatusNotImplemented, map[string]string{
                "error": "Team registration is not implemented",
			})
		} else {
			// Default to Individual
			participantID = userID
		}

		// 5. Check Capacity
		var count int
		countQuery := `SELECT count(*) FROM registrations WHERE tournament_id = $1`
		err = db.QueryRow(context.Background(), countQuery, tournamentID).Scan(&count)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check registration count"})
		}

		if count >= t.MaxParticipants {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Tournament is full"})
		}

		// 6. Insert Registration
		insertQuery := `
			INSERT INTO registrations (tournament_id, participant_id, participant_name, status)
			VALUES ($1, $2, $3, 'approved') 
		`
		// Note: Changed status to 'approved' by default for MVP simplicity, or keep 'pending' if you have approval logic.
		
		_, err = db.Exec(context.Background(), insertQuery, tournamentID, participantID, req.Name)
		if err != nil {
			// Check for Postgres Unique Violation (Error Code 23505)
			if err.Error() == "ERROR: duplicate key value violates unique constraint \"registrations_pkey\" (SQLSTATE 23505)" {
				return c.JSON(http.StatusConflict, map[string]string{"error": "You are already registered"})
			}
			
			log.Printf("Database Insert Error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to register for tournament"})
		}

		return c.JSON(http.StatusCreated, map[string]string{"message": "Successfully registered", "participant_id": participantID})
	}
}

func HealthCheckHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Tournament Service Healthy")
}


// In tournament.go

// Request struct for status updates
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// Currently checks Organizer, but easy to expand for "Judges" or "Admins" later.
func canManageTournament(userID string, t Tournament) bool {
	// Future: Check if userID exists in a 'judges' table for this tournament
	return userID == t.OrganizerID
}

func UpdateTournamentStatusHandler(db *pgxpool.Pool, rmq *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		tournamentID := c.Param("id")
		userID := c.Request().Header.Get("X-User-Id")

		if userID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing authentication"})
		}

		// 1. Bind Request
		var req UpdateStatusRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		// Validate Status Enum (Safety check)
		validStatuses := map[string]bool{
			"draft": true, "registration_open": true, "registration_closed": true,
			"ongoing": true, "completed": true, "cancelled": true,
		}
		if !validStatuses[req.Status] {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid status value"})
		}

		// 2. Fetch Tournament (Need OrganizerID to verify permissions)
		var t Tournament
		query := `SELECT id, organizer_id, status FROM tournaments WHERE id = $1`
		err := db.QueryRow(context.Background(), query, tournamentID).Scan(&t.ID, &t.OrganizerID, &t.Status)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Tournament not found"})
		}

		// 3. Permission Check (Scalable)
		if !canManageTournament(userID, t) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to manage this tournament"})
		}

		// 4. Update Status in DB
		updateQuery := `UPDATE tournaments SET status = $1 WHERE id = $2`
		_, err = db.Exec(context.Background(), updateQuery, req.Status, tournamentID)
		if err != nil {
			log.Printf("Database Update Error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update status"})
		}

		// 5. Publish Event (Crucial for Bracket generation or notifying users)
		// Event: TournamentStatusUpdated
		eventPayload := map[string]string{
			"tournament_id": t.ID,
			"old_status":    t.Status,
			"new_status":    req.Status,
			"updated_by":    userID,
		}
		event := Event{
			EventType: "TournamentStatusUpdated",
			Payload:   eventPayload,
			Timestamp: time.Now(),
		}
		eventBytes, _ := json.Marshal(event)
		_ = rmq.Publish("events.tournament.status_updated", string(eventBytes))

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Tournament status updated successfully", 
			"status": req.Status,
		})
	}
}
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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
func CreateTournamentHandler(db DBClient, rmq EventPublisher) echo.HandlerFunc {
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

type RegistrationRequest struct {
	TeamID string `json:"team_id"` // Optional: Only for team tournaments
	Name   string `json:"name"`    // Required: Display name (User or Team name)
}


func GetAllTournamentsHandler(db DBClient) echo.HandlerFunc {
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

func RegisterTournamentHandler(db DBClient) echo.HandlerFunc {
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
			if req.TeamID == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "This is a team tournament. Team ID is required."})
			}
			participantID = req.TeamID
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

// Now checks for Organizer OR SuperAdmin role
func canManageTournament(userID string, userRoles string, t Tournament) bool {
	// 1. Check for SuperAdmin role
	roles := strings.Split(userRoles, ",")
	for _, role := range roles {
		if strings.TrimSpace(role) == "SuperAdmin" {
			return true
		}
	}

	// 2. Fallback to Organizer check
	return userID == t.OrganizerID
}


func UpdateTournamentStatusHandler(db DBClient, rmq EventPublisher) echo.HandlerFunc {
	return func(c echo.Context) error {
		tournamentID := c.Param("id")
		userID := c.Request().Header.Get("X-User-Id")
		userRoles := c.Request().Header.Get("X-User-Roles") // Get roles

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
		if !canManageTournament(userID, userRoles, t) { // Pass roles
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

func GetTournamentHandler(db DBClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		tournamentID := c.Param("id")
		userID := c.Request().Header.Get("X-User-Id")

		// 1. Fetch Tournament
		var t Tournament
		query := `
			SELECT 
				t.id, t.organizer_id, t.name, COALESCE(t.description, ''), t.game, 
				t.format, t.participant_type, t.start_date, t.status, 
				t.min_participants, t.max_participants, t.public,
				COUNT(r.participant_id) as current_participants
			FROM tournaments t
			LEFT JOIN registrations r ON t.id = r.tournament_id
			WHERE t.id = $1
			GROUP BY t.id
		`
		err := db.QueryRow(context.Background(), query, tournamentID).Scan(
			&t.ID, &t.OrganizerID, &t.Name, &t.Description, &t.Game,
			&t.Format, &t.ParticipantType, &t.StartDate, &t.Status, 
			&t.MinParticipants, &t.MaxParticipants, &t.Public, 
			&t.CurrentParticipants,
		)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Tournament not found"})
		}

		// 2. Security Check
		// If private, ONLY the organizer (or registered users, optional) can see it.
		if !t.Public && t.OrganizerID != userID {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to view this private tournament"})
		}

		return c.JSON(http.StatusOK, t)
	}
}

// Struct for allowed updates (keeps ID/Organizer immutable)
type UpdateTournamentRequest struct {
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Game            string     `json:"game"`
	Format          string     `json:"format"`
	StartDate       *time.Time `json:"start_date"` // Pointer allows checking for null/missing
	Status          string     `json:"status"`
	MinParticipants int        `json:"min_participants"`
	MaxParticipants int        `json:"max_participants"`
	Public          bool       `json:"public"`
}

func UpdateTournamentDetailsHandler(db DBClient, rmq EventPublisher) echo.HandlerFunc {
	return func(c echo.Context) error {
		tournamentID := c.Param("id")
		userID := c.Request().Header.Get("X-User-Id")
		userRoles := c.Request().Header.Get("X-User-Roles")

		if userID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing authentication"})
		}

		// 1. Bind Request
		var req UpdateTournamentRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		// 2. Fetch Existing Data (to check permissions & status)
		var t Tournament
		query := `SELECT id, organizer_id, status FROM tournaments WHERE id = $1`
		err := db.QueryRow(context.Background(), query, tournamentID).Scan(&t.ID, &t.OrganizerID, &t.Status)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Tournament not found"})
		}

		// 3. Permission Check
		if !canManageTournament(userID, userRoles, t) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "You do not have permission to edit this tournament"})
		}

		// 4. Safety Checks (Logic Guard)
		// If tournament is already active/completed, prevent changing Format or Game
		if (t.Status == "ongoing" || t.Status == "completed") && (req.Format != "" || req.Game != "") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Cannot change Game or Format once tournament has started"})
		}

		// 5. Update Query
		// Note: This query updates fields only if they are provided (handling partial updates roughly)
		// For a true PUT, you usually replace all, but here's a robust SQL approach:
		updateQuery := `
			UPDATE tournaments SET
				name = COALESCE(NULLIF($1, ''), name),
				description = $2,
				game = COALESCE(NULLIF($3, ''), game),
				format = COALESCE(NULLIF($4, ''), format),
				start_date = COALESCE($5, start_date),
				status = COALESCE(NULLIF($6, ''), status),
				min_participants = COALESCE(NULLIF($7, 0), min_participants),
				max_participants = COALESCE(NULLIF($8, 0), max_participants),
				public = $9
			WHERE id = $10
		`
		
		_, err = db.Exec(context.Background(), updateQuery,
			req.Name, req.Description, req.Game, req.Format, 
			req.StartDate, req.Status, req.MinParticipants, req.MaxParticipants, req.Public,
			tournamentID,
		)

		if err != nil {
			log.Printf("Update Error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update tournament"})
		}

		// 6. Publish Event
		// Use a lightweight payload or fetch the full updated object
		_ = rmq.Publish("events.tournament.updated", `{"id":"`+tournamentID+`", "action":"details_updated"}`)

		return c.JSON(http.StatusOK, map[string]string{"message": "Tournament updated successfully"})
	}
}

func GetParticipantsHandler(db DBClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		tournamentID := c.Param("id")

		// Query registrations for this tournament
		// You might want to filter by status='approved' if you implement approval logic later
		query := `
			SELECT participant_id, participant_name 
			FROM registrations 
			WHERE tournament_id = $1
		`
		
		rows, err := db.Query(context.Background(), query, tournamentID)
		if err != nil {
			log.Printf("DB Error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch participants"})
		}
		defer rows.Close()

		// Struct matches the JSON expected by Bracket Service
		type Participant struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		
		participants := []Participant{}

		for rows.Next() {
			var p Participant
			if err := rows.Scan(&p.ID, &p.Name); err != nil {
				log.Printf("Scan Error: %v", err)
				continue
			}
			participants = append(participants, p)
		}

		return c.JSON(http.StatusOK, participants)
	}
}
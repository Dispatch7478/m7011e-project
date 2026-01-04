package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

// MockRabbitMQ satisfies the EventPublisher interface
type MockRabbitMQ struct {
	LastKey  string
	LastBody string
	Err      error
}

func (m *MockRabbitMQ) Publish(routingKey string, body string) error {
	m.LastKey = routingKey
	m.LastBody = body
	return m.Err
}

func TestCreateTournamentHandler(t *testing.T) {
	// 1. Setup
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	mockRMQ := &MockRabbitMQ{}

	// 2. Define Request Data
	reqPayload := Tournament{
		Name:            "Test Tournament",
		Description:     "A test description",
		Game:            "Pong",
		Format:          "single-elimination",
		ParticipantType: "individual",
		MinParticipants: 2,
		MaxParticipants: 4,
	}
	body, _ := json.Marshal(reqPayload)

	// 3. Define DB Expectations
	// We expect an INSERT statement.
	// Since ID is generated inside the handler, we use AnyArg() for the first argument ($1).
	mockDB.ExpectExec("INSERT INTO tournaments").
		WithArgs(
			pgxmock.AnyArg(), // ID
			"user-123",       // OrganizerID
			reqPayload.Name,
			reqPayload.Description,
			reqPayload.Game,
			reqPayload.Format,
			reqPayload.ParticipantType,
			pgxmock.AnyArg(), // StartDate (zero value in struct, but matches type)
			"draft",
			reqPayload.MinParticipants,
			reqPayload.MaxParticipants,
			true,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// 4. Create HTTP Request
	req := httptest.NewRequest(http.MethodPost, "/tournaments", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-Id", "user-123") // Mock Authenticated User
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 5. Execute Handler
	handler := CreateTournamentHandler(mockDB, mockRMQ)
	err = handler(c)

	// 6. Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NoError(t, mockDB.ExpectationsWereMet())
	
	// Check RabbitMQ was called
	assert.Equal(t, "events.tournament.created", mockRMQ.LastKey)
	assert.Contains(t, mockRMQ.LastBody, "Test Tournament")
}

func TestGetTournamentHandler_Success(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Mock DB Return Rows
	columns := []string{
		"id", "organizer_id", "name", "description", "game",
		"format", "participant_type", "start_date", "status",
		"min_participants", "max_participants", "public", "current_participants",
	}
	
	// Create a mock row
	mockDB.ExpectQuery("SELECT .* FROM tournaments t").
		WithArgs("tourn-123").
		WillReturnRows(pgxmock.NewRows(columns).AddRow(
			"tourn-123", "user-123", "My Tourney", "Desc", "Pong",
			"single-elimination", "individual", time.Now(), "draft",
			2, 16, true, 5,
		))

	req := httptest.NewRequest(http.MethodGet, "/tournaments/tourn-123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("tourn-123")

	handler := GetTournamentHandler(mockDB)
	_ = handler(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "My Tourney")
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGetTournamentHandler_NotFound(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Mock DB Error
	mockDB.ExpectQuery("SELECT .* FROM tournaments t").
		WithArgs("tourn-999").
		WillReturnError(errors.New("no rows in result set"))

	req := httptest.NewRequest(http.MethodGet, "/tournaments/tourn-999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("tourn-999")

	handler := GetTournamentHandler(mockDB)
	_ = handler(c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}


func TestGetAllTournamentsHandler(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	// 1. Prepare Expected Rows
	// The query selects 13 columns.
	columns := []string{
		"id", "organizer_id", "name", "description", "game",
		"format", "participant_type", "start_date", "status",
		"min_participants", "max_participants", "public", "current_participants",
	}

	mockDB.ExpectQuery("SELECT .* FROM tournaments").
		WillReturnRows(pgxmock.NewRows(columns).
			AddRow("t1", "u1", "Tourney A", "Desc", "Pong", "single", "individual", time.Now(), "open", 2, 8, true, 2).
			AddRow("t2", "u2", "Tourney B", "Desc", "Pong", "single", "individual", time.Now(), "open", 2, 8, true, 0))

	// 2. Execute
	req := httptest.NewRequest(http.MethodGet, "/tournaments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := GetAllTournamentsHandler(mockDB)
	err = handler(c)

	// 3. Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Tourney A")
	assert.Contains(t, rec.Body.String(), "Tourney B")
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestRegisterTournamentHandler_Success(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	// 1. Prepare Mock Data
	tournamentID := "tourn-123"
	userID := "user-100"
	reqBody := `{"name": "Player One"}`

	// Expect transaction
	mockDB.ExpectBegin()

	// 2. Expectation 1: Check Tournament Details
	// Need regex matching (.*) to cover the query string variations
	mockDB.ExpectQuery("SELECT id, status, public, max_participants, participant_type FROM tournaments .* FOR UPDATE").
		WithArgs(tournamentID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "status", "public", "max", "type"}).
			AddRow(tournamentID, "registration_open", true, 16, "individual"))

	// 3. Expectation 2: Check Capacity
	// SELECT count(*) FROM registrations...
	mockDB.ExpectQuery("SELECT count\\(\\*\\) FROM registrations").
		WithArgs(tournamentID).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(5)) // 5/16 slots taken

	// 4. Expectation 3: Insert Registration
	mockDB.ExpectExec("INSERT INTO registrations").
		WithArgs(tournamentID, userID, "Player One").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockDB.ExpectCommit()

	// 5. Execute
	req := httptest.NewRequest(http.MethodPost, "/tournaments/"+tournamentID+"/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-Id", userID)
	req.Header.Set("X-User-Name", "Player One")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tournamentID)

	handler := RegisterTournamentHandler(mockDB)
	err = handler(c)

	// 6. Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestRegisterTournamentHandler_Full(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	tournamentID := "tourn-full"
	userID := "user-101"

	mockDB.ExpectBegin()

	// 1. Tournament is Open...
	mockDB.ExpectQuery("SELECT id, status, public, max_participants, participant_type FROM tournaments").
		WithArgs(tournamentID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "status", "public", "max", "type"}).
			AddRow(tournamentID, "registration_open", true, 16, "individual"))

	// 2. ...But Full (16/16)
	mockDB.ExpectQuery("SELECT count\\(\\*\\) FROM registrations").
		WithArgs(tournamentID).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(16))
	
	mockDB.ExpectRollback()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"Late Player"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-Id", userID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tournamentID)

	handler := RegisterTournamentHandler(mockDB)
	_ = handler(c)

	assert.Equal(t, http.StatusConflict, rec.Code) // Should return 409 Conflict
	assert.Contains(t, rec.Body.String(), "Tournament is full")
}

func TestUpdateTournamentStatusHandler_Success(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()
	mockRMQ := &MockRabbitMQ{}

	tournamentID := "tourn-123"
	organizerID := "user-admin"

	// 1. Fetch Tournament to check permissions
	mockDB.ExpectQuery("SELECT id, organizer_id, status FROM tournaments").
		WithArgs(tournamentID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "organizer_id", "status"}).
			AddRow(tournamentID, organizerID, "draft"))

	// 2. Update Status
	mockDB.ExpectExec("UPDATE tournaments SET status").
		WithArgs("registration_open", tournamentID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{"status":"registration_open"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-Id", organizerID) // Must match organizer
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tournamentID)

	handler := UpdateTournamentStatusHandler(mockDB, mockRMQ)
	_ = handler(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "events.tournament.status_updated", mockRMQ.LastKey)
}

func TestUpdateTournamentDetailsHandler_Success(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()
	mockRMQ := &MockRabbitMQ{}

	tournamentID := "tourn-update"
	organizerID := "user-admin"

	// 1. Fetch Existing Data (Permission Check)
	// Query: SELECT id, organizer_id, status FROM tournaments...
	mockDB.ExpectQuery("SELECT id, organizer_id, status FROM tournaments").
		WithArgs(tournamentID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "organizer_id", "status"}).
			AddRow(tournamentID, organizerID, "draft"))

	// 2. Perform Update
	// Query: UPDATE tournaments SET ...
	mockDB.ExpectExec("UPDATE tournaments SET").
		WithArgs(
			"New Name", "New Desc", pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), true,
			tournamentID,
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// 3. Request
	reqBody := `{"name": "New Name", "description": "New Desc", "public": true}`
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-Id", organizerID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tournamentID)

	handler := UpdateTournamentDetailsHandler(mockDB, mockRMQ)
	_ = handler(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "events.tournament.updated", mockRMQ.LastKey)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestUpdateTournamentDetailsHandler_Locked(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()
	mockRMQ := &MockRabbitMQ{}

	tournamentID := "tourn-ongoing"
	organizerID := "user-admin"

	// 1. Fetch Existing Data
	// Status is "ongoing", which should LOCK Game/Format changes
	mockDB.ExpectQuery("SELECT id, organizer_id, status FROM tournaments").
		WithArgs(tournamentID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "organizer_id", "status"}).
			AddRow(tournamentID, organizerID, "ongoing"))

	// 2. Request try to change Format
	reqBody := `{"format": "double-elimination"}`
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-Id", organizerID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tournamentID)

	handler := UpdateTournamentDetailsHandler(mockDB, mockRMQ)
	_ = handler(c)

	// 3. Expect Conflict (409)
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "Cannot change Game or Format")
}

func TestGetParticipantsHandler(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	tournamentID := "tourn-123"

	// 1. Mock Query
	mockDB.ExpectQuery("SELECT participant_id, participant_name FROM registrations").
		WithArgs(tournamentID).
		WillReturnRows(pgxmock.NewRows([]string{"participant_id", "participant_name"}).
			AddRow("user-1", "Alice").
			AddRow("user-2", "Bob"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tournamentID)

	handler := GetParticipantsHandler(mockDB)
	_ = handler(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Alice")
	assert.Contains(t, rec.Body.String(), "Bob")
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestCreateTournamentHandler_ValidationFailure(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool() // Mock DB is needed to satisfy the handler signature
	assert.NoError(t, err)
	defer mockDB.Close()

	mockRMQ := &MockRabbitMQ{}

	// 1. Invalid Payload (5 participants is not a power of 2)
	reqPayload := Tournament{
		Name:            "Bad Tournament",
		Game:            "Pong",
		ParticipantType: "individual",
		MinParticipants: 2,
		MaxParticipants: 5, // <--- INVALID: Odd number
	}
	body, _ := json.Marshal(reqPayload)

	// 2. no db needed here

	// 3. Request
	req := httptest.NewRequest(http.MethodPost, "/tournaments", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-Id", "user-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := CreateTournamentHandler(mockDB, mockRMQ)
	err = handler(c)

	// 4. Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code) 
	assert.Contains(t, rec.Body.String(), "Max participants must be a multiple of 2")
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestUpdateTournamentStatusHandler_PermissionDenied(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()
	mockRMQ := &MockRabbitMQ{}

	tournamentID := "tourn-123"
	organizerID := "real-organizer"
	attackerID := "hacker"

	// 1. Fetch Tournament
	mockDB.ExpectQuery("SELECT id, organizer_id, status FROM tournaments").
		WithArgs(tournamentID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "organizer_id", "status"}).
			AddRow(tournamentID, organizerID, "draft"))

	// 2. Expect NO Update execution

	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{"status":"registration_open"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-Id", attackerID) // Wrong User
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tournamentID)

	handler := UpdateTournamentStatusHandler(mockDB, mockRMQ)
	_ = handler(c)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGetTournamentHandler_PrivateForbidden(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	tournamentID := "tourn-private"
	organizerID := "vip-user"
	visitorID := "random-user"

	// 1. Fetch returns Public=false
	columns := []string{
		"id", "organizer_id", "name", "description", "game",
		"format", "participant_type", "start_date", "status",
		"min_participants", "max_participants", "public", "current_participants",
	}
	
	mockDB.ExpectQuery("SELECT .* FROM tournaments t").
		WithArgs(tournamentID).
		WillReturnRows(pgxmock.NewRows(columns).AddRow(
			tournamentID, organizerID, "Secret Club", "Desc", "Pong",
			"single", "individual", time.Now(), "draft",
			2, 16, false, 0, // <--- Public is FALSE
		))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Id", visitorID) // Not the organizer
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tournamentID)

	handler := GetTournamentHandler(mockDB)
	_ = handler(c)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "do not have permission")
	assert.NoError(t, mockDB.ExpectationsWereMet())
}
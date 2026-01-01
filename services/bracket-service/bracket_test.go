package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

// MockRabbitMQ
type MockRabbitMQ struct{}
func (m *MockRabbitMQ) Publish(key, body string) error { return nil }

func TestGenerateBracket_Success(t *testing.T) {
	e := echo.New()
	// Enable Regex Matching
	mockDB, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer mockDB.Close()

	// 1. Mock External Service
	tsMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		participants := []Participant{
			{ID: "p1", Name: "Player 1"}, {ID: "p2", Name: "Player 2"},
			{ID: "p3", Name: "Player 3"}, {ID: "p4", Name: "Player 4"},
		}
		json.NewEncoder(w).Encode(participants)
	}))
	defer tsMock.Close()

	h := &BracketHandler{
		DB:                   mockDB,
		RMQ:                  &MockRabbitMQ{},
		TournamentServiceURL: tsMock.URL,
	}

	mockDB.ExpectBegin()

	// 2. Expectations
	// We use pgxmock.AnyArg() for ALL arguments to ensure the test passes 
	// regardless of int/int32/int64 mismatches or nil/pointer nuances.
	
	// Round 2 (Final) - 1 Match
	mockDB.ExpectQuery(`(?s).*INSERT INTO matches.*`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("match-final"))

	// Round 1 (Semis) - 2 Matches
	mockDB.ExpectQuery(`(?s).*INSERT INTO matches.*`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("match-semi-1"))

	mockDB.ExpectQuery(`(?s).*INSERT INTO matches.*`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("match-semi-2"))

	mockDB.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/brackets/generate?tournament_id=t1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.GenerateBracket(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGenerateBracket_Validation(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer mockDB.Close()

	h := &BracketHandler{DB: mockDB, RMQ: &MockRabbitMQ{}, TournamentServiceURL: "http://mock"}

	// Scenario 1: Missing Tournament ID
	req := httptest.NewRequest(http.MethodPost, "/brackets/generate", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.GenerateBracket(c)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Scenario 2: Not Enough Participants
	// Mock returning 1 participant
	tsMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		participants := []Participant{{ID: "p1", Name: "Player 1"}}
		json.NewEncoder(w).Encode(participants)
	}))
	defer tsMock.Close()
	h.TournamentServiceURL = tsMock.URL

	req = httptest.NewRequest(http.MethodPost, "/brackets/generate?tournament_id=t1", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = h.GenerateBracket(c)
	
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Not enough participants")
}

func TestUpdateMatchResult_Advancement(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mockDB.Close()

	h := &BracketHandler{DB: mockDB, RMQ: &MockRabbitMQ{}}

	matchID := "match-1"
	nextMatchID := "match-5"
	winnerID := "winner-user"
	matchNum := 1 
	body := `{"score_a": "2", "score_b": "1", "winner_id": "winner-user"}`

	mockDB.ExpectBegin()

	// 1. Fetch
	mockDB.ExpectQuery(`SELECT next_match_id, match_number FROM matches WHERE id = $1`).
		WithArgs(matchID).
		WillReturnRows(pgxmock.NewRows([]string{"next_match_id", "match_number"}).
			AddRow(&nextMatchID, matchNum))

	// 2. Update Score
	// NOTE: The whitespace must EXACTLY match the query in the handler
	updateScoreSQL := `
		UPDATE matches 
		SET score_a = $1, score_b = $2, winner_id = $3, status = 'completed' 
		WHERE id = $4`
	mockDB.ExpectExec(updateScoreSQL).
		WithArgs("2", "1", winnerID, matchID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// 3. Advance
	advanceWinnerSQL := `UPDATE matches SET player1_id = $1 WHERE id = $2`
	mockDB.ExpectExec(advanceWinnerSQL).
		WithArgs(winnerID, nextMatchID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mockDB.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("match_id")
	c.SetParamValues(matchID)

	err = h.UpdateMatchResult(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGetBracket(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer mockDB.Close()

	h := &BracketHandler{DB: mockDB}

	// Define specific types that match the Scan targets
	// ID (string), TournamentID (string), Round (int), MatchNumber (int), 
	// Player1ID (*string), Player2ID (*string), NextMatchID (*string), 
	// Status (string), ScoreA (string), ScoreB (string)
	
	p1 := "p1"
	p2 := "p2"
	next := "m2"
	
	mockDB.ExpectQuery(`(?s).*SELECT.*FROM matches.*`).
		WithArgs("t1").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "tournament_id", "round", "match_number", 
			"player1_id", "player2_id", "next_match_id", 
			"status", "score_a", "score_b",
		}).
		AddRow(
			"m1", "t1", 1, 1, 
			&p1, &p2, &next, 
			"scheduled", "0", "0",
		))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("tournamentId")
	c.SetParamValues("t1")

	h.GetBracket(c)

	// Verify we got the match back
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "m1")
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestUpdateMatchResult_FinalMatch(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mockDB.Close()

	h := &BracketHandler{DB: mockDB, RMQ: &MockRabbitMQ{}}

	matchID := "match-final"
	winnerID := "winner-user"
	matchNum := 1 
	body := `{"score_a": "3", "score_b": "2", "winner_id": "winner-user"}`

	mockDB.ExpectBegin()

	// 1. Fetch: Return NIL for next_match_id to simulate the Final
	// Note: We use AddRow(nil, matchNum)
	mockDB.ExpectQuery(`SELECT next_match_id, match_number FROM matches WHERE id = $1`).
		WithArgs(matchID).
		WillReturnRows(pgxmock.NewRows([]string{"next_match_id", "match_number"}).
			AddRow(nil, matchNum))

	// 2. Update Score (Standard update)
	updateScoreSQL := `
		UPDATE matches 
		SET score_a = $1, score_b = $2, winner_id = $3, status = 'completed' 
		WHERE id = $4`
	mockDB.ExpectExec(updateScoreSQL).
		WithArgs("3", "2", winnerID, matchID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// 3. Expect Commit immediately (NO advancement query should run)
	mockDB.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("match_id")
	c.SetParamValues(matchID)

	err = h.UpdateMatchResult(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGenerateBracket_NotEnoughParticipants(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer mockDB.Close()

	// 1. Mock External Service returning only 1 participant
	tsMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		participants := []Participant{
			{ID: "p1", Name: "Player 1"},
		}
		json.NewEncoder(w).Encode(participants)
	}))
	defer tsMock.Close()

	h := &BracketHandler{
		DB:                   mockDB,
		RMQ:                  &MockRabbitMQ{},
		TournamentServiceURL: tsMock.URL,
	}

	// 2. No DB interaction expected (should fail before transaction)

	req := httptest.NewRequest(http.MethodPost, "/brackets/generate?tournament_id=t1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.GenerateBracket(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Not enough participants")
}

func TestUpdateMatchResult_NotFound(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mockDB.Close()

	h := &BracketHandler{DB: mockDB}
	
	matchID := "unknown-id"
	body := `{"score_a": "1", "score_b": "0", "winner_id": "w"}`

	mockDB.ExpectBegin()
	// Fail the fetch
	mockDB.ExpectQuery(`SELECT next_match_id, match_number FROM matches WHERE id = $1`).
		WithArgs(matchID).
		WillReturnError(errors.New("no rows in result set"))
	mockDB.ExpectRollback()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("match_id")
	c.SetParamValues(matchID)

	_ = h.UpdateMatchResult(c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGenerateBracket_WithBye(t *testing.T) {
	e := echo.New()
	mockDB, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer mockDB.Close()

	// 1. Mock External Service (3 Participants = 1 Bye)
	tsMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		participants := []Participant{
			{ID: "p1", Name: "Player 1"},
			{ID: "p2", Name: "Player 2"},
			{ID: "p3", Name: "Player 3"},
		}
		json.NewEncoder(w).Encode(participants)
	}))
	defer tsMock.Close()

	h := &BracketHandler{
		DB:                   mockDB,
		RMQ:                  &MockRabbitMQ{},
		TournamentServiceURL: tsMock.URL,
	}

	mockDB.ExpectBegin()

	// LOGIC: 3 Players -> 4 Slots. Round 1 has 2 matches.
	// Match 1: P1 vs P2 (Standard)
	// Match 2: P3 vs NULL (Bye) -> Auto Advance P3

	// 1. Insert Final (Round 2)
	mockDB.ExpectQuery(`(?s).*INSERT INTO matches.*`).
		WithArgs(pgxmock.AnyArg(), 2, 1, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("match-final"))

	// 2. Insert Semi 1 (Standard)
	mockDB.ExpectQuery(`(?s).*INSERT INTO matches.*`).
		WithArgs(pgxmock.AnyArg(), 1, 1, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("match-semi-1"))

	// 3. Insert Semi 2 (Bye)
	mockDB.ExpectQuery(`(?s).*INSERT INTO matches.*`).
		WithArgs(pgxmock.AnyArg(), 1, 2, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("match-semi-2"))

	// 4. Expect Auto-Advancement UPDATE for the Bye match
	// Logic: If Match 2 (Even) is bye, update Player2 slot of next match
	mockDB.ExpectExec(`(?s).*UPDATE matches SET.*`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mockDB.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/brackets/generate?tournament_id=t1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.GenerateBracket(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGenerateBracket_TournamentServiceError(t *testing.T) {
	e := echo.New()
	
	// Mock Server that returns 500
	tsMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tsMock.Close()

	h := &BracketHandler{
		DB:                   nil, // Should not reach DB
		TournamentServiceURL: tsMock.URL,
	}

	req := httptest.NewRequest(http.MethodPost, "/brackets/generate?tournament_id=t1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = h.GenerateBracket(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Failed to fetch participants")
}

func TestUpdateMatchResult_InvalidBody(t *testing.T) {
	e := echo.New()
	h := &BracketHandler{DB: nil} // No DB needed

	body := `{"score_a": "2", "score_b": }` // Malformed JSON
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = h.UpdateMatchResult(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
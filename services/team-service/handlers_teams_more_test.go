package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
)

func TestListTeams_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "name", "tag", "captain_id", "logo_url", "created_at"}).
		AddRow("t1", "Team One", "T1", "u1", nil, now).
		AddRow("t2", "Team Two", "T2", "u2", "http://logo", now)

	mock.ExpectQuery(`SELECT id::text, name, tag, captain_id::text, logo_url, created_at\s+FROM teams`).
		WillReturnRows(rows)

	h := Handler{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/teams", nil)

	h.ListTeams(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var out []Team
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(out))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestMyCaptainTeams_UnauthorizedWithoutUser(t *testing.T) {
	h := Handler{DB: nil}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/teams/captain", nil)

	h.MyCaptainTeams(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestMyMemberTeams_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "name", "tag", "captain_id", "logo_url", "created_at"}).
		AddRow("t1", "Team One", "T1", "u1", nil, now)

	mock.ExpectQuery(`FROM team_members m\s+JOIN teams t ON t\.id = m\.team_id\s+WHERE m\.user_id = \$1::uuid`).
		WithArgs("user-123").
		WillReturnRows(rows)

	h := Handler{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/teams/me", nil).
		WithContext(context.WithValue(context.Background(), ctxUserID, "user-123"))

	h.MyMemberTeams(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var out []Team
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(out) != 1 || out[0].ID != "t1" {
		t.Fatalf("unexpected teams: %#v", out)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestDeleteTeam_NotFoundWhenNoRowsAffected(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT captain_id FROM teams WHERE id = \$1::uuid FOR UPDATE`).
		WithArgs("team-1").
		WillReturnError(sql.ErrNoRows) // Simulate team not found
	mock.ExpectRollback()

	h := Handler{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/teams/team-1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))

	h.DeleteTeam(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestDeleteTeam_NoContentOnSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Mock the full transaction
	mock.ExpectBegin()
	// 1. Check for captain
	mock.ExpectQuery(`SELECT captain_id FROM teams WHERE id = \$1::uuid FOR UPDATE`).
		WithArgs("team-1").
		WillReturnRows(sqlmock.NewRows([]string{"captain_id"}).AddRow("user-123"))
	// 2. Delete invites
	mock.ExpectExec(`DELETE FROM invites WHERE team_id = \$1::uuid`).
		WithArgs("team-1").
		WillReturnResult(sqlmock.NewResult(0, 1)) // Assume 1 invite deleted
	// 3. Delete members
	mock.ExpectExec(`DELETE FROM team_members WHERE team_id = \$1::uuid`).
		WithArgs("team-1").
		WillReturnResult(sqlmock.NewResult(0, 2)) // Assume 2 members deleted
	// 4. Delete team
	mock.ExpectExec(`DELETE FROM teams WHERE id = \$1::uuid`).
		WithArgs("team-1").
		WillReturnResult(sqlmock.NewResult(0, 1)) // Assume 1 team deleted
	mock.ExpectCommit()

	h := Handler{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/teams/team-1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	// This user ID must match the captain_id returned by the SELECT query
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))

	h.DeleteTeam(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestLeaveTeam_BadRequestIfCaptain(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS\(`).
		WithArgs("team-1", "user-123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	h := Handler{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/teams/team-1/leave", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))

	h.LeaveTeam(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rr.Code, rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestLeaveTeam_NoContentOnSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS\(`).
		WithArgs("team-1", "user-123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectExec(`DELETE FROM team_members\s+WHERE team_id = \$1::uuid AND user_id = \$2::uuid`).
		WithArgs("team-1", "user-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	h := Handler{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/teams/team-1/leave", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))

	h.LeaveTeam(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCaptainRemoveMember_NoContentOnSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS\(`).
		WithArgs("team-1", "captain-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectExec(`DELETE FROM team_members\s+WHERE team_id = \$1::uuid AND user_id = \$2::uuid`).
		WithArgs("team-1", "member-9").
		WillReturnResult(sqlmock.NewResult(0, 1))

	h := Handler{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/teams/team-1/members/member-9", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "team-1", "userId": "member-9"})
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "captain-1"))

	h.CaptainRemoveMember(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestIsCaptainOfTeam_ReturnsFalse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS\(`).
		WithArgs("team-1", "user-123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	h := Handler{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/teams/team-1/is-captain", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))

	h.IsCaptainOfTeam(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["is_captain"] != false {
		t.Fatalf("expected is_captain=false, got %#v", out["is_captain"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCreateTeam_InvalidJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	h := Handler{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/teams", bytes.NewBufferString("{bad json"))
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))

	h.CreateTeam(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

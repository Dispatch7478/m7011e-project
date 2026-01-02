package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
)

func TestListTeamMembers_BadRequestWithoutID(t *testing.T) {
	h := Handler{DB: nil}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/teams//members", nil)
	req = mux.SetURLVars(req, map[string]string{"id": ""})

	h.ListTeamMembers(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestListTeamMembers_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	joined := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"user_id", "role", "joined_at"}).
		AddRow("u1", "captain", joined).
		AddRow("u2", "member", joined)

	mock.ExpectQuery(`SELECT user_id::text, role, joined_at\s+FROM team_members`).
		WithArgs("team-1").
		WillReturnRows(rows)

	h := Handler{DB: db}
	req := httptest.NewRequest(http.MethodGet, "/teams/team-1/members", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	rr := httptest.NewRecorder()

	h.ListTeamMembers(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out []TeamMember
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(out) != 2 || out[0].UserID != "u1" {
		t.Fatalf("unexpected out: %#v", out)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAcceptInviteAndJoinTeam_NoContentOnSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT status, expires_at\s+FROM invites`).
		WithArgs("inv-1", "team-1", "x@y.com").
		WillReturnRows(sqlmock.NewRows([]string{"status", "expires_at"}).AddRow("pending", nil))

	mock.ExpectExec(`INSERT INTO team_members`).
		WithArgs("team-1", "user-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`UPDATE invites\s+SET status = 'accepted'`).
		WithArgs("inv-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	h := Handler{DB: db}
	body := []byte(`{"invite_id":"inv-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/teams/team-1/members", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	ctx := context.WithValue(req.Context(), ctxUserID, "user-123")
	ctx = context.WithValue(ctx, ctxEmail, "x@y.com")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.AcceptInviteAndJoinTeam(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rr.Code, rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAcceptInviteAndJoinTeam_GoneIfExpired(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()

	expired := time.Now().Add(-1 * time.Hour)
	mock.ExpectQuery(`SELECT status, expires_at\s+FROM invites`).
		WithArgs("inv-1", "team-1", "x@y.com").
		WillReturnRows(sqlmock.NewRows([]string{"status", "expires_at"}).AddRow("pending", &expired))

	h := Handler{DB: db}
	req := httptest.NewRequest(http.MethodPost, "/teams/team-1/members", bytes.NewBufferString(`{"invite_id":"inv-1"}`))
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	ctx := context.WithValue(req.Context(), ctxUserID, "user-123")
	ctx = context.WithValue(ctx, ctxEmail, "x@y.com")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.AcceptInviteAndJoinTeam(rr, req)

	if rr.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d body=%s", rr.Code, rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

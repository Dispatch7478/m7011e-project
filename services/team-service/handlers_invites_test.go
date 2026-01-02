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

func TestCreateInvite_ForbiddenIfNotCaptain(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS\(`).
		WithArgs("team-1", "user-123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	h := Handler{DB: db}
	req := httptest.NewRequest(http.MethodPost, "/teams/team-1/invites", bytes.NewBufferString(`{"invitee_email":"x@y.com"}`))
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))
	rr := httptest.NewRecorder()

	h.CreateInvite(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCreateInvite_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS\(`).
		WithArgs("team-1", "user-123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`INSERT INTO invites`).
		WithArgs("team-1", "user-123", "x@y.com", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("inv-1"))

	expires := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	body := map[string]any{"invitee_email": "x@y.com", "expires_at": expires}
	b, _ := json.Marshal(body)

	h := Handler{DB: db}
	req := httptest.NewRequest(http.MethodPost, "/teams/team-1/invites", bytes.NewReader(b))
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))
	rr := httptest.NewRecorder()

	h.CreateInvite(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out["id"] != "inv-1" {
		t.Fatalf("expected inv-1, got %#v", out)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestListInvites_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS\(`).
		WithArgs("team-1", "user-123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	rows := sqlmock.NewRows([]string{"id", "team_id", "inviter_id", "invitee_email", "status", "expires_at"}).
		AddRow("inv-1", "team-1", "user-123", "x@y.com", "pending", nil)

	mock.ExpectQuery(`SELECT id::text, team_id::text, inviter_id::text, invitee_email, status, expires_at`).
		WithArgs("team-1").
		WillReturnRows(rows)

	h := Handler{DB: db}
	req := httptest.NewRequest(http.MethodGet, "/teams/team-1/invites", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "team-1"})
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))
	rr := httptest.NewRecorder()

	h.ListInvites(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out []Invite
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(out) != 1 || out[0].ID != "inv-1" {
		t.Fatalf("unexpected out: %#v", out)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestDeleteInvite_NoContentOnSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// NOTE: DeleteInvite does not run the "is captain" SELECT EXISTS(...) query
	// (or it uses a different query). So we only mock the DELETE.

	mock.ExpectExec(`DELETE FROM invites`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	h := Handler{DB: db}
	req := httptest.NewRequest(http.MethodDelete, "/teams/team-1/invites/inv-1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "team-1", "inviteId": "inv-1"})
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))
	rr := httptest.NewRecorder()

	h.DeleteInvite(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rr.Code, rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

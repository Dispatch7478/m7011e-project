package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateTeam_BadRequestWhenMissingFields(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	h := Handler{DB: db}

	body := []byte(`{"name":"","tag":""}`)
	req := httptest.NewRequest(http.MethodPost, "/teams", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))

	rr := httptest.NewRecorder()
	h.CreateTeam(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateTeam_CreatesTeamAndCaptainMember(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	h := Handler{DB: db}

	mock.ExpectBegin()

	// INSERT INTO teams ... RETURNING id
	mock.ExpectQuery(`INSERT INTO teams`).
		WithArgs("My Team", "MT", "user-123", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("team-abc"))

	// INSERT INTO team_members ... (captain)
	mock.ExpectExec(`INSERT INTO team_members`).
		WithArgs("team-abc", "user-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	payload := map[string]any{
		"name":     "My Team",
		"tag":      "MT",
		"logo_url": nil,
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/teams", bytes.NewReader(b))
	req = req.WithContext(context.WithValue(req.Context(), ctxUserID, "user-123"))

	rr := httptest.NewRecorder()
	h.CreateTeam(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rr.Code, rr.Body.String())
	}

	var out map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["id"] != "team-abc" {
		t.Fatalf("expected team id team-abc, got %q", out["id"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

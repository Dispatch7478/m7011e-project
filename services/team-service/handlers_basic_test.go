package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHealth_OK(t *testing.T) {
	h := Handler{DB: nil}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "OK" {
		t.Fatalf("expected body OK, got %q", rr.Body.String())
	}
}

func TestReady_ServiceUnavailableWhenPingFails(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectPing().WillReturnError(assertErr("ping failed"))

	h := Handler{DB: db}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)

	h.Ready(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestReady_OKWhenPingSucceeds(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectPing().WillReturnError(nil)

	h := Handler{DB: db}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)

	h.Ready(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "READY" {
		t.Fatalf("expected READY, got %q", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestTeamsCount_ReturnsJSONCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(int64(7))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM teams;`).WillReturnRows(rows)

	h := Handler{DB: db}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/teams/count", nil)

	h.TeamsCount(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body map[string]int64
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["count"] != 7 {
		t.Fatalf("expected count=7, got %v", body["count"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// tiny helper so we don't import extra packages
type assertErr string

func (e assertErr) Error() string { return string(e) }

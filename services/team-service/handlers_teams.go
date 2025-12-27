package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type createTeamRequest struct {
	Name    string  `json:"name"`
	Tag     string  `json:"tag"`
	LogoURL *string `json:"logo_url"`
}

func (h Handler) ListTeams(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id::text, name, tag, captain_id::text, logo_url, created_at
		FROM teams
		ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Tag, &t.CaptainID, &t.LogoURL, &t.CreatedAt); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		out = append(out, t)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h Handler) MyCaptainTeams(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(`
		SELECT id::text, name, tag, captain_id::text, logo_url, created_at
		FROM teams
		WHERE captain_id = $1::uuid
		ORDER BY created_at DESC`, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Tag, &t.CaptainID, &t.LogoURL, &t.CreatedAt); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		out = append(out, t)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h Handler) MyMemberTeams(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(`
		SELECT t.id::text, t.name, t.tag, t.captain_id::text, t.logo_url, t.created_at
		FROM team_members m
		JOIN teams t ON t.id = m.team_id
		WHERE m.user_id = $1::uuid
		ORDER BY t.created_at DESC`, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Tag, &t.CaptainID, &t.LogoURL, &t.CreatedAt); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		out = append(out, t)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req createTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Tag == "" {
		http.Error(w, "name and tag required", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var teamID string
	err = tx.QueryRow(`
		INSERT INTO teams (name, tag, captain_id, logo_url)
		VALUES ($1, $2, $3::uuid, $4)
		RETURNING id::text`,
		req.Name, req.Tag, userID, req.LogoURL,
	).Scan(&teamID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Ensure captain is also a team member
	_, err = tx.Exec(`
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1::uuid, $2::uuid, 'captain')
		ON CONFLICT DO NOTHING`, teamID, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": teamID})
}

func (h Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	teamID := mux.Vars(r)["id"]
	if teamID == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	// Only captain can delete
	res, err := h.DB.Exec(`
		DELETE FROM teams
		WHERE id = $1::uuid AND captain_id = $2::uuid`,
		teamID, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		// either not found, or not captain
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Optional helper if you later need role checks
func isSQLNoRows(err error) bool { return err == sql.ErrNoRows }

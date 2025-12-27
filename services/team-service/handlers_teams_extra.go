package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type TeamMember struct {
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

func emailFromCtx(r *http.Request) (string, bool) {
	v := r.Context().Value(ctxEmail)
	s, ok := v.(string)
	return s, ok && s != ""
}

func (h Handler) ListTeamMembers(w http.ResponseWriter, r *http.Request) {
	teamID := mux.Vars(r)["id"]
	if teamID == "" {
		http.Error(w, "missing team id", http.StatusBadRequest)
		return
	}

	rows, err := h.DB.Query(`
		SELECT user_id::text, role, joined_at
		FROM team_members
		WHERE team_id = $1::uuid
		ORDER BY
			CASE role
				WHEN 'captain' THEN 0
				WHEN 'admin' THEN 1
				ELSE 2
			END,
			joined_at ASC
	`, teamID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []TeamMember
	for rows.Next() {
		var m TeamMember
		if err := rows.Scan(&m.UserID, &m.Role, &m.JoinedAt); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		out = append(out, m)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type acceptInviteRequest struct {
	InviteID string `json:"invite_id"`
}

// POST /teams/{id}/members  (accept invite)
func (h Handler) AcceptInviteAndJoinTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	email, ok := emailFromCtx(r)
	if !ok {
		http.Error(w, "missing email claim", http.StatusUnauthorized)
		return
	}

	teamID := mux.Vars(r)["id"]
	if teamID == "" {
		http.Error(w, "missing team id", http.StatusBadRequest)
		return
	}

	var req acceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.InviteID == "" {
		http.Error(w, "invite_id required", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Verify invite belongs to this team and is meant for this email and is pending and not expired
	var status string
	var expiresAt *time.Time
	err = tx.QueryRow(`
		SELECT status, expires_at
		FROM invites
		WHERE id = $1::uuid
		  AND team_id = $2::uuid
		  AND lower(invitee_email) = lower($3)
	`, req.InviteID, teamID, email).Scan(&status, &expiresAt)
	if err != nil {
		// Invite not found or email mismatch
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if status != "pending" {
		http.Error(w, "invite not pending", http.StatusConflict)
		return
	}
	if expiresAt != nil && time.Now().After(*expiresAt) {
		http.Error(w, "invite expired", http.StatusGone)
		return
	}

	// Add member (idempotent)
	_, err = tx.Exec(`
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1::uuid, $2::uuid, 'member')
		ON CONFLICT (team_id, user_id) DO NOTHING
	`, teamID, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Mark invite accepted
	_, err = tx.Exec(`
		UPDATE invites
		SET status = 'accepted'
		WHERE id = $1::uuid
	`, req.InviteID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

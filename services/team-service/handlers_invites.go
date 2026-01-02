package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type createInviteRequest struct {
	InviteeEmail string     `json:"invitee_email"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

func (h Handler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	teamID := mux.Vars(r)["id"]
	if teamID == "" {
		http.Error(w, "missing team id", http.StatusBadRequest)
		return
	}

	var req createInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.InviteeEmail == "" {
		http.Error(w, "invitee_email required", http.StatusBadRequest)
		return
	}

	// Simple auth rule for now: only captain can invite
	var exists bool
	if err := h.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM teams WHERE id = $1::uuid AND captain_id = $2::uuid
		)`, teamID, userID).Scan(&exists); err != nil || !exists {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var inviteID string
	err := h.DB.QueryRow(`
		INSERT INTO invites (team_id, inviter_id, invitee_email, status, expires_at)
		VALUES ($1::uuid, $2::uuid, $3, 'pending', $4)
		RETURNING id::text`,
		teamID, userID, req.InviteeEmail, req.ExpiresAt,
	).Scan(&inviteID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": inviteID})
}

func (h Handler) ListInvites(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	teamID := mux.Vars(r)["id"]
	if teamID == "" {
		http.Error(w, "missing team id", http.StatusBadRequest)
		return
	}

	// captain only (for now)
	var exists bool
	if err := h.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM teams WHERE id = $1::uuid AND captain_id = $2::uuid
		)`, teamID, userID).Scan(&exists); err != nil || !exists {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	rows, err := h.DB.Query(`
		SELECT id::text, team_id::text, inviter_id::text, invitee_email, status, expires_at
		FROM invites
		WHERE team_id = $1::uuid and status = 'pending'
		ORDER BY expires_at NULLS LAST`, teamID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []Invite
	for rows.Next() {
		var inv Invite
		if err := rows.Scan(&inv.ID, &inv.TeamID, &inv.InviterID, &inv.InviteeEmail, &inv.Status, &inv.ExpiresAt); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		out = append(out, inv)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h Handler) DeleteInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	email, _ := emailFromCtx(r)

	inviteID := mux.Vars(r)["id"]
	if inviteID == "" {
		http.Error(w, "missing invite id", http.StatusBadRequest)
		return
	}

	// Simple rule: inviter can delete their own invite
	res, err := h.DB.Exec(`
		DELETE FROM invites
		WHERE id = $1::uuid AND (inviter_id = $2::uuid or lower(invitee_email) = lower($3))`,
		inviteID, userID, email)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


type myInviteResponse struct {
	ID           string     `json:"id"`
	TeamID       string     `json:"team_id"`
	TeamName     string     `json:"team_name"` // Needed for UI
	TeamTag      string     `json:"team_tag"`  // Needed for UI
	InviterID    string     `json:"inviter_id"`
	InviteeEmail string     `json:"invitee_email"`
	Status       string     `json:"status"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

func (h Handler) MyInvites(w http.ResponseWriter, r *http.Request) {
	// 1. Get User Email from Context (set by ExtractUser middleware)
	email, ok := emailFromCtx(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. Query invites matching this email + join teams for display info
	rows, err := h.DB.Query(`
		SELECT 
			i.id::text, 
			i.team_id::text, 
			t.name, 
			t.tag, 
			i.inviter_id::text, 
			i.invitee_email, 
			i.status, 
			i.expires_at
		FROM invites i
		JOIN teams t ON i.team_id = t.id
		WHERE LOWER(i.invitee_email) = LOWER($1) 
		  AND i.status = 'pending'
		ORDER BY i.expires_at ASC
	`, email)

	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 3. Scan into response struct
	var out []myInviteResponse
	for rows.Next() {
		var inv myInviteResponse
		if err := rows.Scan(
			&inv.ID, 
			&inv.TeamID, 
			&inv.TeamName, 
			&inv.TeamTag, 
			&inv.InviterID, 
			&inv.InviteeEmail, 
			&inv.Status, 
			&inv.ExpiresAt,
		); err != nil {
			http.Error(w, "db scanning error", http.StatusInternalServerError)
			return
		}
		out = append(out, inv)
	}

	// 4. Return JSON
	w.Header().Set("Content-Type", "application/json")
	if out == nil {
		out = []myInviteResponse{} // Return [] instead of null
	}
	_ = json.NewEncoder(w).Encode(out)
}
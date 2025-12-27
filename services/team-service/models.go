package main

import "time"

type Team struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tag       string    `json:"tag"`
	CaptainID string    `json:"captain_id"`
	LogoURL   *string   `json:"logo_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Invite struct {
	ID           string     `json:"id"`
	TeamID       string     `json:"team_id"`
	InviterID    string     `json:"inviter_id"`
	InviteeEmail string     `json:"invitee_email"`
	Status       string     `json:"status"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

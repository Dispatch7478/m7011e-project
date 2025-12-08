package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	DB *sql.DB
}

// health check
func (h Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Create user
func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Use the ID provided by the api-gateway (from Keycloak)
	query := `INSERT INTO users (id, username, email) VALUES ($1, $2, $3);`

	_, err = h.DB.Exec(query, u.ID, u.Username, u.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// Get User by ID
func (h Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var u User

	query := `SELECT id, username, email, created_at FROM users WHERE id = $1;`
	err := h.DB.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

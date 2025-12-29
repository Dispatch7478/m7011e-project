package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Trigger CI/CD (same idea as your user service)
	fmt.Println("debug build")

	db := InitDB()
	h := Handler{DB: db}

	r := mux.NewRouter()
	r.Use(metricsMiddleware)

	// public endpoints
	r.HandleFunc("/health", h.Health).Methods("GET")
	r.HandleFunc("/ready", h.Ready).Methods("GET")
	r.HandleFunc("/teams/count", h.TeamsCount).Methods("GET")

	// public list for front page
	r.HandleFunc("/teams", h.ListTeams).Methods("GET")
	r.HandleFunc("/teams/{id}/members", h.ListTeamMembers).Methods("GET")

	// auth subrouter
	authed := r.PathPrefix("").Subrouter()
	authed.Use(ExtractUser)
	authed.HandleFunc("/me/teams", h.MyMemberTeams).Methods("GET")
	authed.HandleFunc("/me/teams/captain", h.MyCaptainTeams).Methods("GET")

	authed.HandleFunc("/teams/{id}/leave", h.LeaveTeam).Methods("POST")
	authed.HandleFunc("/teams/{id}/members/{userId}", h.CaptainRemoveMember).Methods("DELETE")

	authed.HandleFunc("/teams", h.CreateTeam).Methods("POST")
	authed.HandleFunc("/teams/{id}", h.DeleteTeam).Methods("DELETE")

	authed.HandleFunc("/teams/{id}/is-captain", h.IsCaptainOfTeam).Methods("GET")

	authed.HandleFunc("/teams/{id}/invites", h.CreateInvite).Methods("POST")
	authed.HandleFunc("/teams/{id}/invites", h.ListInvites).Methods("GET")
	authed.HandleFunc("/invites/{id}", h.DeleteInvite).Methods("DELETE")
	authed.HandleFunc("/teams/{id}/members", h.AcceptInviteAndJoinTeam).Methods("POST")
	// metrics endpoint
	r.Handle("/metrics", metricsHandler()).Methods("GET")

	log.Println("Team Service running on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

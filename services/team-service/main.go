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

	// metrics endpoint
	r.Handle("/metrics", metricsHandler()).Methods("GET")

	log.Println("Team Service running on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

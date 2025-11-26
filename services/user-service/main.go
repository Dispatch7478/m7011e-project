package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("debug build")

	db := InitDB()
	h := Handler{DB: db}

	r := mux.NewRouter()

	// public endpoints
	r.HandleFunc("/health", h.Health).Methods("GET")

	// user endpoints
	r.HandleFunc("/users", h.CreateUser).Methods("POST")
	r.HandleFunc("/users/{id}", h.GetUser).Methods("GET")

	log.Println("User Service running on :8080")
	http.ListenAndServe(":8080", r)
}

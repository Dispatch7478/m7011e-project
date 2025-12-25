package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func InitDB() *sql.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USERNAME")
		pass := os.Getenv("DB_PASSWORD")
		name := os.Getenv("DB_NAME")

		if host == "" || user == "" || pass == "" || name == "" {
			log.Fatal("database config missing: need DB_HOST, DB_USERNAME, DB_PASSWORD, DB_NAME or DB_DSN")
		}

		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user,
			pass,
			host,
			"5432",
			name,
		)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("🚀 Connected to Postgres")
	return db
}

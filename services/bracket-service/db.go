package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBClient interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Close()
}

func ConnectDB() (*pgxpool.Pool, error) {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST") 
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" || dbHost == "" {
		return nil, fmt.Errorf("DB credentials missing")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", dbUser, dbPass, dbHost, dbName)
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	log.Println("Connected to Bracket DB (PostgreSQL)")
	return pool, nil
}
package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool creates and returns a database connection pool
func NewPostgresPool(user, password, host, port, dbName string) *pgxpool.Pool {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("❌ Unable to create database pool: %v", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("❌ Unable to connect to database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL successfully!")
	return pool
}

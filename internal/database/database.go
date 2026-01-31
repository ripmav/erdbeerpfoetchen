package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq" // also for side effects
)

type DB struct {
	conn *sql.DB
}

func Connect(ctx context.Context, uri string) (*DB, error) {
	conn, err := sql.Open("postgres", uri)

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("database connection established")

	conn.SetMaxOpenConns(8)
	conn.SetMaxIdleConns(8)

	conn.SetConnMaxLifetime(3 * time.Minute)

	db := &DB{conn: conn}
	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

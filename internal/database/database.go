package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq" // also for side effects
	"github.com/ripmav/erdbeerpfoetchen/internal/lamimi"
)

type DB struct {
	conn *sql.DB
}

func Connect(ctx context.Context, uri string, debug bool) (*DB, error) {
	if debug {
		uri = fmt.Sprintf("%s?sslmode=disable", uri)
	}

	conn, err := sql.Open("postgres", uri)

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	defer func(conn *sql.DB) {
		_ = conn.Close() // nolint: errcheck
	}(conn)

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

func (db *DB) Update(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return db.transaction(ctx, fn, true)
}

func (db *DB) Read(ctx context.Context, fn func(tx *sql.Tx) error) (*lamimi.Collection, error) {
	return nil, db.transaction(ctx, fn, false)
}

func (db *DB) transaction(ctx context.Context, fn func(tx *sql.Tx) error, write bool) (err error) {
	defer func(start time.Time) {
		took := time.Since(start).Milliseconds()

		if err != nil {
			slog.Error("database transaction failed", "error", err, "write", write, "took", took)
		}
	}(time.Now())

	tx, err := db.conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: !write})

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback() // nolint: errcheck
	}(tx)

	if err := fn(tx); err != nil {
		return fmt.Errorf("execute transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

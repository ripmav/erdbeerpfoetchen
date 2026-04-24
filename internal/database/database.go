package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "github.com/lib/pq" // also for side effects
	"github.com/pressly/goose/v3"

	"github.com/ripmav/erdbeerpfoetchen/schema"
)

type DB struct {
	conn *sql.DB
}

func Connect(ctx context.Context, uri string, debug bool) (*DB, error) {
	if debug {
		uri = strings.TrimRight(uri, "?")
		if strings.Contains(uri, "?") {
			uri = uri + "&sslmode=disable"
		} else {
			uri = uri + "?sslmode=disable"
		}
	}

	conn, err := sql.Open("postgres", uri)

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if debug {
		slog.InfoContext(ctx, "database connection established with debug mode enabled sslmode=disable")
	} else {
		slog.InfoContext(ctx, "database connection established")
	}

	conn.SetMaxOpenConns(8)
	conn.SetMaxIdleConns(8)

	conn.SetConnMaxLifetime(3 * time.Minute)

	db := &DB{conn: conn}
	return db, nil
}

func ConnectAndMigrate(ctx context.Context, uri string, debug bool) (*DB, error) {
	db, err := Connect(ctx, uri, debug)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Migrate(ctx context.Context) error {
	goose.SetBaseFS(schema.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db.conn, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func (db *DB) Update(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return db.transaction(ctx, fn, true)
}

func (db *DB) Read(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return db.transaction(ctx, fn, false)
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

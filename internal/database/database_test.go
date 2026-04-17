package database_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/ripmav/erdbeerpfoetchen/internal/database"
)

// testDB opens a real database connection using TEST_DB_URI and skips the test
// if the variable is not set.
func testDB(t *testing.T) *database.DB {
	t.Helper()
	uri := os.Getenv("TEST_DB_URI")
	if uri == "" {
		t.Skip("TEST_DB_URI not set — skipping integration test")
	}
	db, err := database.Connect(context.Background(), uri, false)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestConnect(t *testing.T) {
	testDB(t) // passes if Connect succeeds
}

func TestConnect_DebugMode(t *testing.T) {
	uri := os.Getenv("TEST_DB_URI")
	if uri == "" {
		t.Skip("TEST_DB_URI not set — skipping integration test")
	}
	db, err := database.Connect(context.Background(), uri, true)
	if err != nil {
		t.Fatalf("Connect debug=true: %v", err)
	}
	_ = db.Close()
}

func TestConnect_InvalidHost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := database.Connect(ctx, "postgres://invalid-host-xyz:5432/nope", false)
	if err == nil {
		t.Error("expected error for unreachable host, got nil")
	}
}

func TestDB_Close(t *testing.T) {
	db := testDB(t)
	if err := db.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestDB_Update_NoOp(t *testing.T) {
	db := testDB(t)
	err := db.Update(context.Background(), func(_ *sql.Tx) error { return nil })
	if err != nil {
		t.Errorf("Update no-op: %v", err)
	}
}

func TestDB_Read_NoOp(t *testing.T) {
	db := testDB(t)
	err := db.Read(context.Background(), func(_ *sql.Tx) error { return nil })
	if err != nil {
		t.Errorf("Read no-op: %v", err)
	}
}

func TestDB_Update_FnError(t *testing.T) {
	db := testDB(t)
	fnErr := errors.New("fn error")
	err := db.Update(context.Background(), func(_ *sql.Tx) error { return fnErr })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, fnErr) {
		t.Errorf("errors.Is mismatch: got %v", err)
	}
}

func TestDB_Read_FnError(t *testing.T) {
	db := testDB(t)
	fnErr := errors.New("fn error")
	err := db.Read(context.Background(), func(_ *sql.Tx) error { return fnErr })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, fnErr) {
		t.Errorf("errors.Is mismatch: got %v", err)
	}
}

func TestDB_Update_ContextCancelled(t *testing.T) {
	db := testDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before use
	err := db.Update(ctx, func(_ *sql.Tx) error { return nil })
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

func TestNewCollectionRepository(t *testing.T) {
	db := testDB(t)
	repo := database.NewCollectionRepository(db, false)
	if repo == nil {
		t.Fatal("NewCollectionRepository returned nil")
	}
}

func TestNewUserRepository(t *testing.T) {
	db := testDB(t)
	repo := database.NewUserRepository(db, false)
	if repo == nil {
		t.Fatal("NewUserRepository returned nil")
	}
}

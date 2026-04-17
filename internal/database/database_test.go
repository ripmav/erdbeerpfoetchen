package database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/ripmav/erdbeerpfoetchen/internal/database"
)

// initSQL contains the DDL statements required to set up the streaming schema
// (mirrors the goose migrations in schema/migrations/).
const initSQL = `
CREATE SCHEMA "streaming";

CREATE TABLE "streaming"."user" (
    "id"         uuid    NOT NULL DEFAULT gen_random_uuid(),
    "user_name"  varchar NOT NULL,
    "api_token"  uuid    NOT NULL DEFAULT gen_random_uuid(),
    PRIMARY KEY ("id")
);

CREATE TABLE "streaming"."collection" (
    "collection_key" varchar NOT NULL,
    "streamer"       uuid    NOT NULL,
    "viewer"         varchar NOT NULL,
    "json"           jsonb   NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY ("collection_key", "streamer", "viewer"),
    CONSTRAINT user_id_collection_fk
        FOREIGN KEY ("streamer")
        REFERENCES "streaming"."user" ("id")
        ON DELETE CASCADE
);
`

// startPostgres spins up a real PostgreSQL container, runs the schema
// migrations, and returns the connection URI. The container is terminated
// automatically when the test finishes.
func startPostgres(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err, "failed to start postgres container")
	t.Cleanup(func() { _ = ctr.Terminate(context.Background()) })

	uri, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	// Apply schema migrations directly.
	rawDB, err := sql.Open("postgres", uri)
	require.NoError(t, err)
	defer rawDB.Close()

	_, err = rawDB.ExecContext(ctx, initSQL)
	require.NoError(t, err, "failed to apply schema migrations")

	return uri
}

// testDB opens a *database.DB backed by the test container.
func testDB(t *testing.T) *database.DB {
	t.Helper()
	uri := startPostgres(t)
	db, err := database.Connect(context.Background(), uri, false)
	require.NoError(t, err, "Connect")
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestConnect(t *testing.T) {
	testDB(t) // passes if Connect succeeds without error
}

func TestConnect_DebugMode(t *testing.T) {
	ctx := context.Background()
	ctr, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ctr.Terminate(context.Background()) })

	// Get URI without sslmode so Connect can append it in debug mode.
	uri, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := database.Connect(ctx, uri, true)
	require.NoError(t, err, "Connect debug=true")
	_ = db.Close()
}

func TestConnect_InvalidHost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := database.Connect(ctx, "postgres://invalid-host-xyz:5432/nope?sslmode=disable", false)
	assert.Error(t, err, "expected error for unreachable host")
}

func TestDB_Close(t *testing.T) {
	db := testDB(t)
	err := db.Close()
	assert.NoError(t, err, "Close")
}

func TestDB_Update_NoOp(t *testing.T) {
	db := testDB(t)
	err := db.Update(context.Background(), func(_ *sql.Tx) error { return nil })
	assert.NoError(t, err, "Update no-op")
}

func TestDB_Read_NoOp(t *testing.T) {
	db := testDB(t)
	err := db.Read(context.Background(), func(_ *sql.Tx) error { return nil })
	assert.NoError(t, err, "Read no-op")
}

func TestDB_Update_FnError(t *testing.T) {
	db := testDB(t)
	fnErr := errors.New("fn error")
	err := db.Update(context.Background(), func(_ *sql.Tx) error { return fnErr })
	require.Error(t, err)
	assert.ErrorIs(t, err, fnErr)
}

func TestDB_Read_FnError(t *testing.T) {
	db := testDB(t)
	fnErr := errors.New("fn error")
	err := db.Read(context.Background(), func(_ *sql.Tx) error { return fnErr })
	require.Error(t, err)
	assert.ErrorIs(t, err, fnErr)
}

func TestDB_Update_ContextCancelled(t *testing.T) {
	db := testDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before use
	err := db.Update(ctx, func(_ *sql.Tx) error { return nil })
	assert.Error(t, err, "expected error for cancelled context")
}

func TestNewCollectionRepository(t *testing.T) {
	db := testDB(t)
	repo := database.NewCollectionRepository(db, false)
	require.NotNil(t, repo, "NewCollectionRepository returned nil")
}

func TestNewUserRepository(t *testing.T) {
	db := testDB(t)
	repo := database.NewUserRepository(db, false)
	require.NotNil(t, repo, "NewUserRepository returned nil")
}

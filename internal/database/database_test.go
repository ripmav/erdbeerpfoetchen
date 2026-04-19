package database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ripmav/erdbeerpfoetchen/internal/database"
	"github.com/ripmav/erdbeerpfoetchen/internal/sqltest"
)

func TestConnect_InvalidHost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := database.Connect(ctx, "postgres://invalid-host-xyz:5432/nope?sslmode=disable", false)
	assert.Error(t, err, "expected error for unreachable host")
}

func TestDB_Close(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	err = container.Destroy(t.Context())
	assert.NoError(t, err, "Close")
}

func TestDB_Update_NoOp(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.Connect(t.Context(), container.URI, true)
	assert.NoError(t, err)
	err = db.Update(context.Background(), func(_ *sql.Tx) error { return nil })
	assert.NoError(t, err, "Update no-op")
}

func TestDB_Read_NoOp(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.Connect(t.Context(), container.URI, true)
	assert.NoError(t, err)
	err = db.Read(context.Background(), func(_ *sql.Tx) error { return nil })
	assert.NoError(t, err, "Read no-op")
}

func TestDB_Update_FnError(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.Connect(t.Context(), container.URI, true)
	assert.NoError(t, err)
	fnErr := errors.New("fn error")
	err = db.Update(context.Background(), func(_ *sql.Tx) error { return fnErr })
	require.Error(t, err)
	assert.ErrorIs(t, err, fnErr)
}

func TestDB_Read_FnError(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.Connect(t.Context(), container.URI, true)
	assert.NoError(t, err)
	fnErr := errors.New("fn error")
	err = db.Read(context.Background(), func(_ *sql.Tx) error { return fnErr })
	require.Error(t, err)
	assert.ErrorIs(t, err, fnErr)
}

func TestDB_Update_ContextCancelled(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.Connect(t.Context(), container.URI, true)
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before use
	err = db.Update(ctx, func(_ *sql.Tx) error { return nil })
	assert.Error(t, err, "expected error for cancelled context")
}

func TestNewCollectionRepository(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.Connect(t.Context(), container.URI, true)
	assert.NoError(t, err)
	repo := database.NewCollectionRepository(db, false)
	require.NotNil(t, repo, "NewCollectionRepository returned nil")
}

func TestNewUserRepository(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.Connect(t.Context(), container.URI, true)
	assert.NoError(t, err)
	repo := database.NewUserRepository(db, false)
	require.NotNil(t, repo, "NewUserRepository returned nil")
}

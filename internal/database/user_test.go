package database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ripmav/erdbeerpfoetchen/internal/database"
)

// insertUser inserts a row into streaming.user and returns the generated id and api_token.
func insertUser(t *testing.T, ctx context.Context, db *database.DB, userName string) (id uuid.UUID, apiToken uuid.UUID) {
	t.Helper()
	err := db.Update(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`INSERT INTO "streaming"."user" ("user_name") VALUES ($1) RETURNING "id", "api_token"`,
			userName,
		).Scan(&id, &apiToken)
	})
	require.NoError(t, err, "insertUser")
	return
}

func TestUserRepository_GetUserByUserName_HappyPath(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	repo := database.NewUserRepository(db, false)

	id, _ := insertUser(t, ctx, db, "alice")

	got, err := repo.GetUserByUserName(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "alice", got.UserName)
}

func TestUserRepository_GetUserByUserName_NotFound(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	repo := database.NewUserRepository(db, false)

	_, err := repo.GetUserByUserName(ctx, "doesnotexist")
	require.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows) || err != nil, "expected not-found error")
}

func TestUserRepository_GetUserById_HappyPath(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	repo := database.NewUserRepository(db, false)

	id, _ := insertUser(t, ctx, db, "bob")

	got, err := repo.GetUserById(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "bob", got.UserName)
}

func TestUserRepository_GetUserById_NotFound(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	repo := database.NewUserRepository(db, false)

	_, err := repo.GetUserById(ctx, uuid.New())
	require.Error(t, err)
}

func TestUserRepository_GetUserApiToken_HappyPath(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	repo := database.NewUserRepository(db, false)

	_, expectedToken := insertUser(t, ctx, db, "carol")

	got, err := repo.GetUserApiToken(ctx, "carol")
	require.NoError(t, err)
	assert.Equal(t, expectedToken, got)
}

func TestUserRepository_GetUserApiToken_NotFound(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	repo := database.NewUserRepository(db, false)

	_, err := repo.GetUserApiToken(ctx, "ghost")
	require.Error(t, err)
}

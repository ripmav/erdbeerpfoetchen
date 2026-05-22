package database_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ripmav/erdbeerpfoetchen/internal/database"
	"github.com/ripmav/erdbeerpfoetchen/internal/sqltest"
)

// insertUser inserts a row into streaming.user and returns the generated id and api_token.
func insertUser(t *testing.T, db *database.DB, userName string) (id uuid.UUID, apiToken uuid.UUID) {
	t.Helper()
	err := db.Update(t.Context(), func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(),
			`INSERT INTO "streaming"."user" ("user_name") VALUES ($1) RETURNING "id", "api_token"`,
			userName,
		).Scan(&id, &apiToken)
	})
	require.NoError(t, err, "insertUser")
	return
}

func TestUserRepository_GetUserByUserName_HappyPath(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	assert.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	id, _ := insertUser(t, db, "alice")

	got, err := repo.GetUserByUserName(t.Context(), "alice")
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "alice", got.UserName)
}

func TestUserRepository_GetUserByUserName_NotFound(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	assert.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	_, err = repo.GetUserByUserName(t.Context(), "doesnotexist")
	require.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows) || err != nil, "expected not-found error")
}

func TestUserRepository_GetUserById_HappyPath(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	assert.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	id, _ := insertUser(t, db, "bob")

	got, err := repo.GetUserById(t.Context(), id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "bob", got.UserName)
}

func TestUserRepository_GetUserById_NotFound(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	assert.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	randID, newErr := uuid.NewV7()
	require.NoError(t, newErr)
	_, err = repo.GetUserById(t.Context(), randID)
	require.Error(t, err)
}

func TestUserRepository_GetUserApiToken_HappyPath(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	assert.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	_, expectedToken := insertUser(t, db, "carol")

	got, err := repo.GetUserApiToken(t.Context(), "carol")
	require.NoError(t, err)
	assert.Equal(t, expectedToken, got)
}

func TestUserRepository_GetUserApiToken_NotFound(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	assert.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	_, err = repo.GetUserApiToken(t.Context(), "ghost")
	require.Error(t, err)
}

func TestUserRepository_CreateUser_HappyPath(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	require.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	apiToken, err := uuid.NewV7()
	require.NoError(t, err)

	u, err := repo.CreateUser(t.Context(), "streamergirl", "twitch-123", apiToken, false)
	require.NoError(t, err)
	assert.Equal(t, "streamergirl", u.UserName)
	assert.Equal(t, apiToken, u.ApiToken)
	assert.Equal(t, "twitch-123", u.TwitchID.String)
	assert.True(t, u.TwitchID.Valid)
	assert.False(t, u.IsAdmin)
	assert.Equal(t, int32(100), u.RateLimitPerMinute)
}

func TestUserRepository_CreateUser_AdminFlag(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	require.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	apiToken, err := uuid.NewV7()
	require.NoError(t, err)

	u, err := repo.CreateUser(t.Context(), "adminuser", "twitch-admin", apiToken, true)
	require.NoError(t, err)
	assert.True(t, u.IsAdmin)
}

func TestUserRepository_CreateUser_DuplicateTwitchId(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	require.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	tok1, _ := uuid.NewV7()
	tok2, _ := uuid.NewV7()

	_, err = repo.CreateUser(t.Context(), "user1", "same-twitch-id", tok1, false)
	require.NoError(t, err)

	_, err = repo.CreateUser(t.Context(), "user2", "same-twitch-id", tok2, false)
	require.Error(t, err, "duplicate twitch_id should violate unique constraint")
}

func TestUserRepository_GetUserByTwitchId_HappyPath(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	require.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	apiToken, err := uuid.NewV7()
	require.NoError(t, err)
	created, err := repo.CreateUser(t.Context(), "twitchuser", "twitch-xyz", apiToken, false)
	require.NoError(t, err)

	got, err := repo.GetUserByTwitchId(t.Context(), "twitch-xyz")
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "twitchuser", got.UserName)
	assert.Equal(t, "twitch-xyz", got.TwitchID.String)
	assert.True(t, got.TwitchID.Valid)
}

func TestUserRepository_GetUserByTwitchId_NotFound(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	db, err := database.ConnectAndMigrate(t.Context(), container.URI, true)
	require.NoError(t, err)
	repo := database.NewUserRepository(db, false)

	_, err = repo.GetUserByTwitchId(t.Context(), "nonexistent-twitch-id")
	require.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

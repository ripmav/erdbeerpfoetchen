package database_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ripmav/erdbeerpfoetchen/internal/collection"
	"github.com/ripmav/erdbeerpfoetchen/internal/database"
)

func TestCollectionRepository_WriteCollection(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	repo := database.NewCollectionRepository(db, false)

	streamerID, _ := insertUser(t, ctx, db, "streamer1")

	raw, _ := json.Marshal(map[string]string{"foo": "bar"})
	col := &collection.Collection{RawMessage: raw}

	err := repo.WriteCollection(ctx, "key1", "viewer1", streamerID, col)
	require.NoError(t, err)
}

func TestCollectionRepository_WriteCollection_Upsert(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	repo := database.NewCollectionRepository(db, false)

	streamerID, _ := insertUser(t, ctx, db, "streamer2")

	raw1, _ := json.Marshal(map[string]string{"v": "1"})
	raw2, _ := json.Marshal(map[string]string{"v": "2"})

	err := repo.WriteCollection(ctx, "key2", "viewer2", streamerID, &collection.Collection{RawMessage: raw1})
	require.NoError(t, err)

	err = repo.WriteCollection(ctx, "key2", "viewer2", streamerID, &collection.Collection{RawMessage: raw2})
	require.NoError(t, err)

	got, err := repo.ReadCollection(ctx, "key2", "viewer2", streamerID)
	require.NoError(t, err)
	assert.JSONEq(t, string(raw2), string(got.RawMessage))
}

func TestCollectionRepository_ReadCollection(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	repo := database.NewCollectionRepository(db, false)

	streamerID, _ := insertUser(t, ctx, db, "streamer3")

	raw, _ := json.Marshal(map[string]string{"hello": "world"})
	col := &collection.Collection{RawMessage: raw}

	require.NoError(t, repo.WriteCollection(ctx, "key3", "viewer3", streamerID, col))

	got, err := repo.ReadCollection(ctx, "key3", "viewer3", streamerID)
	require.NoError(t, err)
	assert.JSONEq(t, string(raw), string(got.RawMessage))
}

func TestCollectionRepository_ReadCollection_NotFound(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	repo := database.NewCollectionRepository(db, false)

	_, err := repo.ReadCollection(ctx, "missing", "nobody", uuid.New())
	require.Error(t, err)
}

package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ripmav/erdbeerpfoetchen/internal/collection"
	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"

	"github.com/google/uuid"
)

type CollectionRepository struct {
	db *DB
}

var _ collection.Repository = (*CollectionRepository)(nil)

func NewCollectionRepository(db *DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

func (repo *CollectionRepository) WriteCollection(ctx context.Context, key, user string, streamerId uuid.UUID, collection *collection.Collection) error {
	return repo.db.Update(ctx, func(tx *sql.Tx) error {
		q := model.New(tx)

		writeParams := model.UpsertCollectionParams{
			Key:      key,
			Viewer:   user,
			Streamer: streamerId,
			Json:     collection.RawMessage,
		}

		if err := q.UpsertCollection(ctx, writeParams); err != nil {
			return fmt.Errorf("failed to upsert collection collection: %w", err)
		}

		return nil
	})
}

func (repo *CollectionRepository) ReadCollection(ctx context.Context, key, user string, streamerId uuid.UUID) (*collection.Collection, error) {
	c := &collection.Collection{}
	err := repo.db.Read(ctx, func(tx *sql.Tx) error {
		q := model.New(tx)

		readParams := model.GetCollectionParams{
			Key:      key,
			UserHash: user,
			Streamer: streamerId,
		}

		sc, err := q.GetCollection(ctx, readParams)

		if err != nil {
			return fmt.Errorf("failed to get collection collection: %w", err)
		}

		c.RawMessage = sc.Json

		return nil
	})

	if err != nil {
		return nil, err
	}

	return c, nil
}

package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ripmav/erdbeerpfoetchen/internal/collection"
	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"

	"github.com/google/uuid"
)

type LamimiRepository struct {
	db *DB
}

var _ collection.Repository = (*LamimiRepository)(nil)

func NewCollectionRepository(db *DB) *LamimiRepository {
	return &LamimiRepository{db: db}
}

func (repo *LamimiRepository) WriteCollection(ctx context.Context, key, user string, streamerId uuid.UUID, collection *collection.Collection) error {
	return repo.db.Update(ctx, func(tx *sql.Tx) error {
		q := model.New(tx)

		writeParams := model.UpsertLamimiCollectionParams{
			CollectionKey:      key,
			CollectionUser:     user,
			CollectionStreamer: streamerId,
			CollectionJson:     collection.RawMessage,
		}

		if err := q.UpsertLamimiCollection(ctx, writeParams); err != nil {
			return fmt.Errorf("failed to upsert collection collection: %w", err)
		}

		return nil
	})
}

func (repo *LamimiRepository) ReadCollection(ctx context.Context, key, user string, streamerId uuid.UUID) (*collection.Collection, error) {
	collection := &collection.Collection{}
	err := repo.db.Read(ctx, func(tx *sql.Tx) error {
		q := model.New(tx)

		readParams := model.GetLamimiCollectionParams{
			CollectionKey:      key,
			CollectionUser:     user,
			CollectionStreamer: streamerId,
		}

		c, err := q.GetLamimiCollection(ctx, readParams)

		if err != nil {
			return fmt.Errorf("failed to get collection collection: %w", err)
		}

		collection.RawMessage = c.CollectionJson

		return nil
	})

	if err != nil {
		return nil, err
	}

	return collection, nil
}

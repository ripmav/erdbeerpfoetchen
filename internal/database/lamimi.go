package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"
	"github.com/ripmav/erdbeerpfoetchen/internal/lamimi"
)

type LamimiRepository struct {
	db *DB
}

var _ lamimi.Repository = (*LamimiRepository)(nil)

func NewLamimiRepository(db *DB) *LamimiRepository {
	return &LamimiRepository{db: db}
}

func (repo *LamimiRepository) WriteCollection(ctx context.Context, key string, user string, collection *lamimi.Collection) error {
	return repo.db.Update(ctx, func(tx *sql.Tx) error {
		q := model.New(tx)

		writeParams := model.UpsertLamimiCollectionParams{
			CollectionKey:  key,
			CollectionUser: user,
			CollectionJson: collection.RawMessage,
		}

		if err := q.UpsertLamimiCollection(ctx, writeParams); err != nil {
			return fmt.Errorf("failed to upsert lamimi collection: %w", err)
		}

		return nil
	})
}

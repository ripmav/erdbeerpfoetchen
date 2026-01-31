package database

import (
	"context"

	"github.com/ripmav/erdbeerpfoetchen/internal/lamimi"
)

type LamimiRepository struct {
	db *DB
}

var _ lamimi.Repository = (*LamimiRepository)(nil)

func NewLamimiRepository(db *DB) *LamimiRepository {
	return &LamimiRepository{db: db}
}

func (l *LamimiRepository) WriteCollection(ctx context.Context, key string, collection *lamimi.Collection) error {
	// TODO implement me
	panic("implement me")
}

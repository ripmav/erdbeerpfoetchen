package database

import (
	"context"
	"database/sql"

	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"
	"github.com/ripmav/erdbeerpfoetchen/internal/streamer"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *DB
}

var _ streamer.Repository = (*UserRepository)(nil)

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (repo *UserRepository) GetUser(ctx context.Context, streamerName string) (uuid.UUID, error) {
	streamerId := uuid.Nil
	err := repo.db.Read(ctx, func(tx *sql.Tx) (err error) {
		q := model.New(tx)

		streamerId, err = q.GetUserApiToken(ctx, streamerName)

		return
	})

	if err != nil {
		return uuid.Nil, err
	}

	return streamerId, nil
}

package database

import (
	"context"
	"database/sql"

	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"
	"github.com/ripmav/erdbeerpfoetchen/internal/user"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *DB
}

var _ user.Repository = (*UserRepository)(nil)

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (repo *UserRepository) GetUserApiToken(ctx context.Context, userName string) (uuid.UUID, error) {
	streamerId := uuid.Nil
	err := repo.db.Read(ctx, func(tx *sql.Tx) (err error) {
		q := model.New(tx)

		streamerId, err = q.GetUserApiToken(ctx, userName)

		return
	})

	if err != nil {
		return uuid.Nil, err
	}

	return streamerId, nil
}

func (repo *UserRepository) GetUserById(ctx context.Context, userId uuid.UUID) (*model.StreamingUser, error) {
	u := model.StreamingUser{ID: userId}
	err := repo.db.Read(ctx, func(tx *sql.Tx) (err error) {
		q := model.New(tx)

		u, err = q.GetUserById(ctx, u.ID)

		return
	})

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (repo *UserRepository) GetUserByUserName(ctx context.Context, userName string) (*model.StreamingUser, error) {
	u := model.StreamingUser{UserName: userName}
	err := repo.db.Read(ctx, func(tx *sql.Tx) (err error) {
		q := model.New(tx)

		u, err = q.GetUserByUserName(ctx, u.UserName)

		return
	})

	if err != nil {
		return nil, err
	}

	return &u, nil
}

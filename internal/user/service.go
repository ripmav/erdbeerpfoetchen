package user

import (
	"context"
	"fmt"

	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"

	"github.com/google/uuid"
)

type Repository interface {
	GetUserApiToken(ctx context.Context, userName string) (uuid.UUID, error)
	GetUserById(ctx context.Context, userId uuid.UUID) (*model.StreamingUser, error)
	GetUserByUserName(ctx context.Context, userName string) (*model.StreamingUser, error)
}

type Service struct {
	userRepo Repository
}

func New(userRepo Repository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) GetUserByUserName(ctx context.Context, userName string) (*model.StreamingUser, error) {
	user, err := s.userRepo.GetUserByUserName(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("cannot get user by username: %w", err)
	}
	return user, nil
}

func (s *Service) GetUserById(ctx context.Context, userId uuid.UUID) (*model.StreamingUser, error) {
	user, err := s.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot get user by id: %w", err)
	}
	return user, nil
}

func (s *Service) GetUserApiToken(ctx context.Context, userName string) (uuid.UUID, error) {
	apiToken, err := s.userRepo.GetUserApiToken(ctx, userName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot get user api token: %w", err)
	}
	return apiToken, nil
}

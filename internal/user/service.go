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
	GetUserRateLimit(ctx context.Context, userName string) (int32, error)
	GetUserByTwitchId(ctx context.Context, twitchID string) (*model.StreamingUser, error)
	CreateUser(ctx context.Context, userName, twitchID string, apiToken uuid.UUID, isAdmin bool) (*model.StreamingUser, error)
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

func (s *Service) GetUserRateLimit(ctx context.Context, userName string) (int32, error) {
	limit, err := s.userRepo.GetUserRateLimit(ctx, userName)
	if err != nil {
		return 0, fmt.Errorf("cannot get user rate limit: %w", err)
	}
	return limit, nil
}

func (s *Service) GetUserByTwitchId(ctx context.Context, twitchID string) (*model.StreamingUser, error) {
	u, err := s.userRepo.GetUserByTwitchId(ctx, twitchID)
	if err != nil {
		return nil, fmt.Errorf("cannot get user by twitch id: %w", err)
	}
	return u, nil
}

func (s *Service) CreateUser(ctx context.Context, userName, twitchID string, isAdmin bool) (*model.StreamingUser, error) {
	apiToken, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("cannot generate api token: %w", err)
	}
	u, err := s.userRepo.CreateUser(ctx, userName, twitchID, apiToken, isAdmin)
	if err != nil {
		return nil, fmt.Errorf("cannot create user: %w", err)
	}
	return u, nil
}

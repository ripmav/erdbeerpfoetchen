package lamimi

import (
	"context"
	"fmt"
)

type Repository interface {
	WriteCollection(ctx context.Context, key, streamerId, user string, collection *Collection) error
	ReadCollection(ctx context.Context, key, streamerId, user string) (*Collection, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) WriteLamimiCollections(ctx context.Context, key, streamerId, user string, lamimi *Collection) error {
	if err := s.repo.WriteCollection(ctx, key, streamerId, user, lamimi); err != nil {
		return fmt.Errorf("cannot write collection: %w", err)
	}

	return nil
}

func (s *Service) ReadLamimiCollections(ctx context.Context, key, streamerId, user string) (*Collection, error) {
	collection, err := s.repo.ReadCollection(ctx, key, streamerId, user)
	if err != nil {
		return nil, fmt.Errorf("cannot read collection: %w", err)
	}

	return collection, nil
}

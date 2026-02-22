package collection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	WriteCollection(ctx context.Context, key, user string, streamerId uuid.UUID, collection *Collection) error
	ReadCollection(ctx context.Context, key, user string, streamerId uuid.UUID) (*Collection, error)
}

type Service struct {
	collectionRepo Repository
}

func New(collectionRepo Repository) *Service {
	return &Service{
		collectionRepo: collectionRepo,
	}
}

func (s *Service) WriteCollection(ctx context.Context, key, user string, streamerId uuid.UUID, lamimi *Collection) error {
	if err := s.collectionRepo.WriteCollection(ctx, key, user, streamerId, lamimi); err != nil {
		return fmt.Errorf("cannot write collection: %w", err)
	}

	return nil
}

func (s *Service) ReadCollection(ctx context.Context, key, user string, streamerId uuid.UUID) (*Collection, error) {
	collection, err := s.collectionRepo.ReadCollection(ctx, key, user, streamerId)
	if err != nil {
		return nil, fmt.Errorf("cannot read collection: %w", err)
	}

	return collection, nil
}

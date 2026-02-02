package lamimi

import (
	"context"
	"encoding/json"
	"fmt"
)

type Collection struct {
	json.RawMessage
}

type Repository interface {
	WriteCollection(ctx context.Context, key, user string, collection *Collection) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) WriteLamimiCollections(ctx context.Context, key, user string, lamimi *Collection) error {
	if err := s.repo.WriteCollection(ctx, key, user, lamimi); err != nil {
		return fmt.Errorf("cannot write collection: %w", err)
	}

	return nil
}

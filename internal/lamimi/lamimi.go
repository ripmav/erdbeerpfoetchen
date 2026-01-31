package lamimi

import (
	"context"
	"encoding/json"
)

type Collection struct {
	json.RawMessage
}

type Repository interface {
	WriteCollection(ctx context.Context, collection *Collection) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) WriteCollection(ctx context.Context, collection *Collection) error {
	if err := s.repo.WriteCollection(ctx, collection); err != nil {
		return err
	}
	return nil
}

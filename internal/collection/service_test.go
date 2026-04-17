package collection_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/ripmav/erdbeerpfoetchen/internal/collection"
)

type mockRepo struct {
	writeErr   error
	readResult *collection.Collection
	readErr    error
}

func (m *mockRepo) WriteCollection(_ context.Context, _, _ string, _ uuid.UUID, _ *collection.Collection) error {
	return m.writeErr
}

func (m *mockRepo) ReadCollection(_ context.Context, _, _ string, _ uuid.UUID) (*collection.Collection, error) {
	return m.readResult, m.readErr
}

func TestNew(t *testing.T) {
	if collection.New(&mockRepo{}) == nil {
		t.Fatal("New returned nil")
	}
}

func TestService_WriteCollection(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	c := &collection.Collection{RawMessage: []byte(`{}`)}

	t.Run("delegates to repo successfully", func(t *testing.T) {
		svc := collection.New(&mockRepo{})
		if err := svc.WriteCollection(ctx, "key", "user", id, c); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("wraps repo error with prefix", func(t *testing.T) {
		repoErr := errors.New("db error")
		svc := collection.New(&mockRepo{writeErr: repoErr})
		err := svc.WriteCollection(ctx, "key", "user", id, c)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repoErr) {
			t.Errorf("errors.Is mismatch: got %v", err)
		}
		if !strings.HasPrefix(err.Error(), "cannot write collection:") {
			t.Errorf("expected prefix 'cannot write collection:', got %q", err.Error())
		}
	})
}

func TestService_ReadCollection(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	t.Run("returns collection from repo", func(t *testing.T) {
		want := &collection.Collection{RawMessage: []byte(`{"x":1}`)}
		svc := collection.New(&mockRepo{readResult: want})
		got, err := svc.ReadCollection(ctx, "key", "user", id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got.RawMessage) != string(want.RawMessage) {
			t.Errorf("got %s, want %s", got.RawMessage, want.RawMessage)
		}
	})

	t.Run("returns nil collection on repo error", func(t *testing.T) {
		repoErr := errors.New("db error")
		svc := collection.New(&mockRepo{readErr: repoErr})
		got, err := svc.ReadCollection(ctx, "key", "user", id)
		if got != nil {
			t.Errorf("expected nil collection, got %v", got)
		}
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repoErr) {
			t.Errorf("errors.Is mismatch: got %v", err)
		}
		if !strings.HasPrefix(err.Error(), "cannot read collection:") {
			t.Errorf("expected prefix 'cannot read collection:', got %q", err.Error())
		}
	})
}

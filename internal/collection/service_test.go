package collection_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	svc := collection.New(&mockRepo{})
	require.NotNil(t, svc, "New returned nil")
}

func TestService_WriteCollection(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	c := &collection.Collection{RawMessage: []byte(`{}`)}

	t.Run("delegates to repo successfully", func(t *testing.T) {
		svc := collection.New(&mockRepo{})
		err := svc.WriteCollection(ctx, "key", "user", id, c)
		assert.NoError(t, err)
	})

	t.Run("wraps repo error with prefix", func(t *testing.T) {
		repoErr := errors.New("db error")
		svc := collection.New(&mockRepo{writeErr: repoErr})
		err := svc.WriteCollection(ctx, "key", "user", id, c)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Contains(t, err.Error(), "cannot write collection:")
	})
}

func TestService_ReadCollection(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	t.Run("returns collection from repo", func(t *testing.T) {
		want := &collection.Collection{RawMessage: []byte(`{"x":1}`)}
		svc := collection.New(&mockRepo{readResult: want})
		got, err := svc.ReadCollection(ctx, "key", "user", id)
		require.NoError(t, err)
		assert.Equal(t, string(want.RawMessage), string(got.RawMessage))
	})

	t.Run("returns nil collection on repo error", func(t *testing.T) {
		repoErr := errors.New("db error")
		svc := collection.New(&mockRepo{readErr: repoErr})
		got, err := svc.ReadCollection(ctx, "key", "user", id)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Contains(t, err.Error(), "cannot read collection:")
	})
}

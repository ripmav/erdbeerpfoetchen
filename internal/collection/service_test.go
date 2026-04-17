package collection_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ripmav/erdbeerpfoetchen/internal/collection"
	"github.com/ripmav/erdbeerpfoetchen/internal/collection/mock"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := collection.New(mock.NewMockRepository(ctrl))
	require.NotNil(t, svc, "New returned nil")
}

func TestService_WriteCollection(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	c := &collection.Collection{RawMessage: []byte(`{}`)}

	t.Run("delegates to repo successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mock.NewMockRepository(ctrl)
		repo.EXPECT().WriteCollection(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		svc := collection.New(repo)
		err := svc.WriteCollection(ctx, "key", "user", id, c)
		assert.NoError(t, err)
	})

	t.Run("wraps repo error with prefix", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mock.NewMockRepository(ctrl)
		repoErr := errors.New("db error")
		repo.EXPECT().WriteCollection(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(repoErr)
		svc := collection.New(repo)
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
		ctrl := gomock.NewController(t)
		repo := mock.NewMockRepository(ctrl)
		want := &collection.Collection{RawMessage: []byte(`{"x":1}`)}
		repo.EXPECT().ReadCollection(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(want, nil)
		svc := collection.New(repo)
		got, err := svc.ReadCollection(ctx, "key", "user", id)
		require.NoError(t, err)
		assert.Equal(t, string(want.RawMessage), string(got.RawMessage))
	})

	t.Run("returns nil collection on repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mock.NewMockRepository(ctrl)
		repoErr := errors.New("db error")
		repo.EXPECT().ReadCollection(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, repoErr)
		svc := collection.New(repo)
		got, err := svc.ReadCollection(ctx, "key", "user", id)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Contains(t, err.Error(), "cannot read collection:")
	})
}

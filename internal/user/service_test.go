package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"
	"github.com/ripmav/erdbeerpfoetchen/internal/user"
	"github.com/ripmav/erdbeerpfoetchen/internal/user/mock"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := user.New(mock.NewMockRepository(ctrl))
	require.NotNil(t, svc, "New returned nil")
}

func TestService_GetUserByUserName(t *testing.T) {
	ctx := context.Background()
	want := &model.StreamingUser{UserName: "alice"}

	t.Run("returns user from repo", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mock.NewMockRepository(ctrl)
		repo.EXPECT().GetUserByUserName(gomock.Any(), gomock.Any()).Return(want, nil)
		svc := user.New(repo)
		got, err := svc.GetUserByUserName(ctx, "alice")
		require.NoError(t, err)
		assert.Equal(t, want.UserName, got.UserName)
	})

	t.Run("wraps repo error with prefix", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mock.NewMockRepository(ctrl)
		repoErr := errors.New("db error")
		repo.EXPECT().GetUserByUserName(gomock.Any(), gomock.Any()).Return(nil, repoErr)
		svc := user.New(repo)
		_, err := svc.GetUserByUserName(ctx, "alice")
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Contains(t, err.Error(), "cannot get user by username:")
	})
}

func TestService_GetUserById(t *testing.T) {
	ctx := context.Background()
	id, err := uuid.NewV7()
	require.NoError(t, err)
	want := &model.StreamingUser{ID: id}

	t.Run("returns user from repo", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mock.NewMockRepository(ctrl)
		repo.EXPECT().GetUserById(gomock.Any(), gomock.Any()).Return(want, nil)
		svc := user.New(repo)
		got, err := svc.GetUserById(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, want.ID, got.ID)
	})

	t.Run("returns nil and wrapped error on failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mock.NewMockRepository(ctrl)
		repoErr := errors.New("db error")
		repo.EXPECT().GetUserById(gomock.Any(), gomock.Any()).Return(nil, repoErr)
		svc := user.New(repo)
		got, err := svc.GetUserById(ctx, id)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Contains(t, err.Error(), "cannot get user by id:")
	})
}

func TestService_GetUserApiToken(t *testing.T) {
	ctx := context.Background()
	wantToken := uuid.New()

	t.Run("returns token from repo", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mock.NewMockRepository(ctrl)
		repo.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(wantToken, nil)
		svc := user.New(repo)
		got, err := svc.GetUserApiToken(ctx, "alice")
		require.NoError(t, err)
		assert.Equal(t, wantToken, got)
	})

	t.Run("returns uuid.Nil and wrapped error on failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mock.NewMockRepository(ctrl)
		repoErr := errors.New("db error")
		repo.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(uuid.Nil, repoErr)
		svc := user.New(repo)
		got, err := svc.GetUserApiToken(ctx, "alice")
		assert.Equal(t, uuid.Nil, got)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Contains(t, err.Error(), "cannot get user api token:")
	})
}

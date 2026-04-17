package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"
	"github.com/ripmav/erdbeerpfoetchen/internal/user"
)

type mockRepo struct {
	apiToken      uuid.UUID
	apiTokenErr   error
	userByID      *model.StreamingUser
	userByIDErr   error
	userByName    *model.StreamingUser
	userByNameErr error
}

func (m *mockRepo) GetUserApiToken(_ context.Context, _ string) (uuid.UUID, error) {
	return m.apiToken, m.apiTokenErr
}

func (m *mockRepo) GetUserById(_ context.Context, _ uuid.UUID) (*model.StreamingUser, error) {
	return m.userByID, m.userByIDErr
}

func (m *mockRepo) GetUserByUserName(_ context.Context, _ string) (*model.StreamingUser, error) {
	return m.userByName, m.userByNameErr
}

func TestNew(t *testing.T) {
	svc := user.New(&mockRepo{})
	require.NotNil(t, svc, "New returned nil")
}

func TestService_GetUserByUserName(t *testing.T) {
	ctx := context.Background()
	want := &model.StreamingUser{UserName: "alice"}

	t.Run("returns user from repo", func(t *testing.T) {
		svc := user.New(&mockRepo{userByName: want})
		got, err := svc.GetUserByUserName(ctx, "alice")
		require.NoError(t, err)
		assert.Equal(t, want.UserName, got.UserName)
	})

	t.Run("wraps repo error with prefix", func(t *testing.T) {
		repoErr := errors.New("db error")
		svc := user.New(&mockRepo{userByNameErr: repoErr})
		_, err := svc.GetUserByUserName(ctx, "alice")
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Contains(t, err.Error(), "cannot get user by username:")
	})
}

func TestService_GetUserById(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	want := &model.StreamingUser{ID: id}

	t.Run("returns user from repo", func(t *testing.T) {
		svc := user.New(&mockRepo{userByID: want})
		got, err := svc.GetUserById(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, want.ID, got.ID)
	})

	t.Run("returns nil and wrapped error on failure", func(t *testing.T) {
		repoErr := errors.New("db error")
		svc := user.New(&mockRepo{userByIDErr: repoErr})
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
		svc := user.New(&mockRepo{apiToken: wantToken})
		got, err := svc.GetUserApiToken(ctx, "alice")
		require.NoError(t, err)
		assert.Equal(t, wantToken, got)
	})

	t.Run("returns uuid.Nil and wrapped error on failure", func(t *testing.T) {
		repoErr := errors.New("db error")
		svc := user.New(&mockRepo{apiTokenErr: repoErr})
		got, err := svc.GetUserApiToken(ctx, "alice")
		assert.Equal(t, uuid.Nil, got)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
		assert.Contains(t, err.Error(), "cannot get user api token:")
	})
}

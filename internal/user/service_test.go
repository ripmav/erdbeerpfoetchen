package user_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
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
	if user.New(&mockRepo{}) == nil {
		t.Fatal("New returned nil")
	}
}

func TestService_GetUserByUserName(t *testing.T) {
	ctx := context.Background()
	want := &model.StreamingUser{UserName: "alice"}

	t.Run("returns user from repo", func(t *testing.T) {
		svc := user.New(&mockRepo{userByName: want})
		got, err := svc.GetUserByUserName(ctx, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.UserName != want.UserName {
			t.Errorf("got %q, want %q", got.UserName, want.UserName)
		}
	})

	t.Run("wraps repo error with prefix", func(t *testing.T) {
		repoErr := errors.New("db error")
		svc := user.New(&mockRepo{userByNameErr: repoErr})
		_, err := svc.GetUserByUserName(ctx, "alice")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repoErr) {
			t.Errorf("errors.Is mismatch: got %v", err)
		}
		if !strings.HasPrefix(err.Error(), "cannot get user by username:") {
			t.Errorf("expected prefix 'cannot get user by username:', got %q", err.Error())
		}
	})
}

func TestService_GetUserById(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	want := &model.StreamingUser{ID: id}

	t.Run("returns user from repo", func(t *testing.T) {
		svc := user.New(&mockRepo{userByID: want})
		got, err := svc.GetUserById(ctx, id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != want.ID {
			t.Errorf("got %v, want %v", got.ID, want.ID)
		}
	})

	t.Run("returns nil and wrapped error on failure", func(t *testing.T) {
		repoErr := errors.New("db error")
		svc := user.New(&mockRepo{userByIDErr: repoErr})
		got, err := svc.GetUserById(ctx, id)
		if got != nil {
			t.Errorf("expected nil user, got %v", got)
		}
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repoErr) {
			t.Errorf("errors.Is mismatch: got %v", err)
		}
		if !strings.HasPrefix(err.Error(), "cannot get user by id:") {
			t.Errorf("expected prefix 'cannot get user by id:', got %q", err.Error())
		}
	})
}

func TestService_GetUserApiToken(t *testing.T) {
	ctx := context.Background()
	wantToken := uuid.New()

	t.Run("returns token from repo", func(t *testing.T) {
		svc := user.New(&mockRepo{apiToken: wantToken})
		got, err := svc.GetUserApiToken(ctx, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != wantToken {
			t.Errorf("got %v, want %v", got, wantToken)
		}
	})

	t.Run("returns uuid.Nil and wrapped error on failure", func(t *testing.T) {
		repoErr := errors.New("db error")
		svc := user.New(&mockRepo{apiTokenErr: repoErr})
		got, err := svc.GetUserApiToken(ctx, "alice")
		if got != uuid.Nil {
			t.Errorf("expected uuid.Nil, got %v", got)
		}
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, repoErr) {
			t.Errorf("errors.Is mismatch: got %v", err)
		}
		if !strings.HasPrefix(err.Error(), "cannot get user api token:") {
			t.Errorf("expected prefix 'cannot get user api token:', got %q", err.Error())
		}
	})
}

package handle_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/ripmav/erdbeerpfoetchen/internal/collection"
	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"
	"github.com/ripmav/erdbeerpfoetchen/internal/handle"
)

const validAuthHeader = "Bearer my-super-secret-key"

type mockCollectionService struct {
	writeErr   error
	readResult *collection.Collection
	readErr    error
}

func (m *mockCollectionService) WriteCollection(_ context.Context, _, _ string, _ uuid.UUID, _ *collection.Collection) error {
	return m.writeErr
}

func (m *mockCollectionService) ReadCollection(_ context.Context, _, _ string, _ uuid.UUID) (*collection.Collection, error) {
	return m.readResult, m.readErr
}

type mockUserService struct {
	apiToken    uuid.UUID
	apiTokenErr error
	user        *model.StreamingUser
	userErr     error
}

func (m *mockUserService) GetUserApiToken(_ context.Context, _ string) (uuid.UUID, error) {
	return m.apiToken, m.apiTokenErr
}

func (m *mockUserService) GetUserByUserName(_ context.Context, _ string) (*model.StreamingUser, error) {
	return m.user, m.userErr
}

func installMux(svc *mockCollectionService, usr *mockUserService) *http.ServeMux {
	mux := http.NewServeMux()
	handle.NewCollectionHandler(svc, usr).Install(mux)
	return mux
}

func TestNewCollectionHandler(t *testing.T) {
	h := handle.NewCollectionHandler(&mockCollectionService{}, &mockUserService{})
	if h == nil {
		t.Fatal("NewCollectionHandler returned nil")
	}
}

func TestCollectionHandler_Install(t *testing.T) {
	mux := installMux(&mockCollectionService{}, &mockUserService{})
	// Both routes should be registered; a request to an unregistered path gets 404.
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("got %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestCollectionHandler_Write(t *testing.T) {
	streamerID := uuid.New()

	post := func(mux *http.ServeMux, body string, headers map[string]string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/collection/alice", strings.NewReader(body))
		req.Header.Set("Authorization", validAuthHeader)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		return rr
	}

	bothKeys := map[string]string{"X-USER-KEY": "user1", "X-COLLECTION-KEY": "col1"}

	t.Run("missing X-USER-KEY returns 400", func(t *testing.T) {
		mux := installMux(&mockCollectionService{}, &mockUserService{apiToken: streamerID})
		rr := post(mux, `{}`, map[string]string{"X-COLLECTION-KEY": "col1"})
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing X-COLLECTION-KEY returns 400", func(t *testing.T) {
		mux := installMux(&mockCollectionService{}, &mockUserService{apiToken: streamerID})
		rr := post(mux, `{}`, map[string]string{"X-USER-KEY": "user1"})
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		mux := installMux(&mockCollectionService{}, &mockUserService{apiToken: streamerID})
		rr := post(mux, `not-json`, bothKeys)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("GetUserApiToken error returns 500", func(t *testing.T) {
		usr := &mockUserService{apiTokenErr: errors.New("db error")}
		mux := installMux(&mockCollectionService{}, usr)
		rr := post(mux, `{}`, bothKeys)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("got %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("WriteCollection error returns 500", func(t *testing.T) {
		svc := &mockCollectionService{writeErr: errors.New("write error")}
		mux := installMux(svc, &mockUserService{apiToken: streamerID})
		rr := post(mux, `{}`, bothKeys)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("got %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success returns 202", func(t *testing.T) {
		mux := installMux(&mockCollectionService{}, &mockUserService{apiToken: streamerID})
		rr := post(mux, `{}`, bothKeys)
		if rr.Code != http.StatusAccepted {
			t.Errorf("got %d, want %d", rr.Code, http.StatusAccepted)
		}
	})
}

func TestCollectionHandler_Read(t *testing.T) {
	streamerID := uuid.New()
	rawJSON := json.RawMessage(`{"foo":"bar"}`)
	c := &collection.Collection{RawMessage: rawJSON}
	u := &model.StreamingUser{ID: streamerID}

	get := func(mux *http.ServeMux, headers map[string]string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
		req.Header.Set("Authorization", validAuthHeader)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		return rr
	}

	bothKeys := map[string]string{"X-USER-KEY": "user1", "X-COLLECTION-KEY": "col1"}

	t.Run("missing X-USER-KEY returns 400", func(t *testing.T) {
		mux := installMux(&mockCollectionService{readResult: c}, &mockUserService{user: u})
		rr := get(mux, map[string]string{"X-COLLECTION-KEY": "col1"})
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing X-COLLECTION-KEY returns 400", func(t *testing.T) {
		mux := installMux(&mockCollectionService{readResult: c}, &mockUserService{user: u})
		rr := get(mux, map[string]string{"X-USER-KEY": "user1"})
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("GetUserByUserName error returns 500", func(t *testing.T) {
		usr := &mockUserService{userErr: errors.New("db error")}
		mux := installMux(&mockCollectionService{readResult: c}, usr)
		rr := get(mux, bothKeys)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("got %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("ReadCollection error returns 500", func(t *testing.T) {
		svc := &mockCollectionService{readErr: errors.New("read error")}
		mux := installMux(svc, &mockUserService{user: u})
		rr := get(mux, bothKeys)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("got %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success returns 200 with JSON body and headers", func(t *testing.T) {
		mux := installMux(&mockCollectionService{readResult: c}, &mockUserService{user: u})
		rr := get(mux, bothKeys)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type: got %q, want %q", ct, "application/json")
		}
		if body := rr.Body.String(); body != string(rawJSON) {
			t.Errorf("body: got %q, want %q", body, string(rawJSON))
		}
	})
}

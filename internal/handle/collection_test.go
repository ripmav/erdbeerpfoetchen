package handle_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ripmav/erdbeerpfoetchen/internal/collection"
	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"
	"github.com/ripmav/erdbeerpfoetchen/internal/handle"
	"github.com/ripmav/erdbeerpfoetchen/internal/handle/mock"
	"github.com/ripmav/erdbeerpfoetchen/internal/middleware"
)

// testToken is a fixed UUID used as the bearer token in test requests.
var testToken = uuid.MustParse("11111111-1111-1111-1111-111111111111")

const validAuthHeader = "Bearer 11111111-1111-1111-1111-111111111111"

func installMux(svc handle.CollectionService, usr handle.UserService) *http.ServeMux {
	mux := http.NewServeMux()
	rl := middleware.NewRateLimiter(nil)
	handle.NewCollectionHandler(svc, usr, rl).Install(mux)
	return mux
}

func TestNewCollectionHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	rl := middleware.NewRateLimiter(nil)
	h := handle.NewCollectionHandler(mock.NewMockCollectionService(ctrl), mock.NewMockUserService(ctrl), rl)
	require.NotNil(t, h, "NewCollectionHandler returned nil")
}

func TestCollectionHandler_Install(t *testing.T) {
	ctrl := gomock.NewController(t)
	mux := installMux(mock.NewMockCollectionService(ctrl), mock.NewMockUserService(ctrl))
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCollectionHandler_Write(t *testing.T) {
	streamerID, err := uuid.NewV7()
	require.NoError(t, err)

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
		ctrl := gomock.NewController(t)
		usr := mock.NewMockUserService(ctrl)
		usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil) // auth
		mux := installMux(mock.NewMockCollectionService(ctrl), usr)
		rr := post(mux, `{}`, map[string]string{"X-COLLECTION-KEY": "col1"})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing X-COLLECTION-KEY returns 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		usr := mock.NewMockUserService(ctrl)
		usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil) // auth
		mux := installMux(mock.NewMockCollectionService(ctrl), usr)
		rr := post(mux, `{}`, map[string]string{"X-USER-KEY": "user1"})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		usr := mock.NewMockUserService(ctrl)
		usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil) // auth
		mux := installMux(mock.NewMockCollectionService(ctrl), usr)
		rr := post(mux, `not-json`, bothKeys)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("GetUserApiToken error returns 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		svc := mock.NewMockCollectionService(ctrl)
		usr := mock.NewMockUserService(ctrl)
		gomock.InOrder(
			usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil),                   // auth
			usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(uuid.Nil, errors.New("db error")), // handler
		)
		mux := installMux(svc, usr)
		rr := post(mux, `{}`, bothKeys)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("WriteCollection error returns 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		svc := mock.NewMockCollectionService(ctrl)
		usr := mock.NewMockUserService(ctrl)
		gomock.InOrder(
			usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil),  // auth
			usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(streamerID, nil), // handler
		)
		svc.EXPECT().WriteCollection(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("write error"))
		mux := installMux(svc, usr)
		rr := post(mux, `{}`, bothKeys)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("success returns 202", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		svc := mock.NewMockCollectionService(ctrl)
		usr := mock.NewMockUserService(ctrl)
		gomock.InOrder(
			usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil),  // auth
			usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(streamerID, nil), // handler
		)
		svc.EXPECT().WriteCollection(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mux := installMux(svc, usr)
		rr := post(mux, `{}`, bothKeys)
		assert.Equal(t, http.StatusAccepted, rr.Code)
	})
}

func TestCollectionHandler_Read(t *testing.T) {
	streamerID, err := uuid.NewV7()
	require.NoError(t, err)
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
		ctrl := gomock.NewController(t)
		usr := mock.NewMockUserService(ctrl)
		usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil) // auth
		mux := installMux(mock.NewMockCollectionService(ctrl), usr)
		rr := get(mux, map[string]string{"X-COLLECTION-KEY": "col1"})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing X-COLLECTION-KEY returns 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		usr := mock.NewMockUserService(ctrl)
		usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil) // auth
		mux := installMux(mock.NewMockCollectionService(ctrl), usr)
		rr := get(mux, map[string]string{"X-USER-KEY": "user1"})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("GetUserByUserName error returns 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		svc := mock.NewMockCollectionService(ctrl)
		usr := mock.NewMockUserService(ctrl)
		usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil) // auth
		usr.EXPECT().GetUserByUserName(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
		mux := installMux(svc, usr)
		rr := get(mux, bothKeys)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("ReadCollection error returns 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		svc := mock.NewMockCollectionService(ctrl)
		usr := mock.NewMockUserService(ctrl)
		usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil) // auth
		usr.EXPECT().GetUserByUserName(gomock.Any(), gomock.Any()).Return(u, nil)
		svc.EXPECT().ReadCollection(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("read error"))
		mux := installMux(svc, usr)
		rr := get(mux, bothKeys)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("success returns 200 with JSON body and headers", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		svc := mock.NewMockCollectionService(ctrl)
		usr := mock.NewMockUserService(ctrl)
		usr.EXPECT().GetUserApiToken(gomock.Any(), gomock.Any()).Return(testToken, nil) // auth
		usr.EXPECT().GetUserByUserName(gomock.Any(), gomock.Any()).Return(u, nil)
		svc.EXPECT().ReadCollection(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(c, nil)
		mux := installMux(svc, usr)
		rr := get(mux, bothKeys)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.Equal(t, string(rawJSON), rr.Body.String())
	})
}

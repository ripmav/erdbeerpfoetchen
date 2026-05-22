package handle

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"

	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"
	"github.com/ripmav/erdbeerpfoetchen/internal/handle/mock"
)

// buildAuthMux wires an AuthHandler pointing its OAuth token exchange and
// Twitch user-info calls at the given test-server URLs.
func buildAuthMux(t *testing.T, svc *mock.MockAuthUserService, tokenURL, usersURL string, adminIDs []string) *http.ServeMux {
	t.Helper()
	h := newAuthHandler(&oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost/auth/twitch/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://unused.invalid",
			TokenURL: tokenURL,
		},
	}, adminIDs, svc)
	h.usersURL = usersURL
	mux := http.NewServeMux()
	h.Install(mux)
	return mux
}

// captureStateCookie calls GET /auth/twitch and returns the state cookie.
func captureStateCookie(t *testing.T, mux *http.ServeMux) *http.Cookie {
	t.Helper()
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/auth/twitch", nil))
	require.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	for _, c := range rr.Result().Cookies() {
		if c.Name == stateCookieName {
			return c
		}
	}
	t.Fatal("state cookie not set after GET /auth/twitch")
	return nil
}

// callbackReq builds a GET /auth/twitch/callback request carrying the given state cookie.
func callbackReq(cookie *http.Cookie, code string) *http.Request {
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/auth/twitch/callback?code=%s&state=%s", code, cookie.Value), nil)
	req.AddCookie(cookie)
	return req
}

// mockTokenSrv returns a test server that emits a minimal OAuth2 token response.
func mockTokenSrv(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-access-token",
			"token_type":   "bearer",
			"expires_in":   3600,
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// mockUsersSrv returns a test server that emits a Twitch /helix/users response.
func mockUsersSrv(t *testing.T, twitchID, login string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{{"id": twitchID, "login": login}},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// --- Login page ---

func TestAuthHandler_LoginPage_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	mux := buildAuthMux(t, mock.NewMockAuthUserService(ctrl), "", "", nil)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, rr.Body.String(), "Login with Twitch")
}

func TestAuthHandler_LoginPage_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mux := buildAuthMux(t, mock.NewMockAuthUserService(ctrl), "", "", nil)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/unknown-path", nil))
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- Login redirect ---

func TestAuthHandler_Login_Redirects(t *testing.T) {
	ctrl := gomock.NewController(t)
	mux := buildAuthMux(t, mock.NewMockAuthUserService(ctrl), "", "", nil)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/auth/twitch", nil))

	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("Location"))
}

func TestAuthHandler_Login_SetsStateCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	mux := buildAuthMux(t, mock.NewMockAuthUserService(ctrl), "", "", nil)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/auth/twitch", nil))

	var found bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == stateCookieName {
			found = true
			assert.NotEmpty(t, c.Value)
			assert.True(t, c.HttpOnly)
			assert.False(t, c.Secure, "test handler uses HTTP; Secure must be false")
		}
	}
	assert.True(t, found, "expected state cookie")
}

func TestAuthHandler_Login_SetsSecureCookieWhenSecure(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newAuthHandler(&oauth2.Config{
		ClientID: "test-client", ClientSecret: "test-secret",
		RedirectURL: "http://localhost/auth/twitch/callback",
		Endpoint:    oauth2.Endpoint{AuthURL: "https://unused.invalid", TokenURL: "https://unused.invalid"},
	}, nil, mock.NewMockAuthUserService(ctrl))
	h.secure = true
	mux := http.NewServeMux()
	h.Install(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/auth/twitch", nil))

	var found bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == stateCookieName {
			found = true
			assert.True(t, c.Secure)
		}
	}
	assert.True(t, found, "expected state cookie")
}

// --- Callback: early-exit error paths (no OAuth servers needed) ---

func TestAuthHandler_Callback_MissingStateCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	mux := buildAuthMux(t, mock.NewMockAuthUserService(ctrl), "", "", nil)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/auth/twitch/callback?code=x&state=y", nil))

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Authentication failed")
}

func TestAuthHandler_Callback_StateMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	mux := buildAuthMux(t, mock.NewMockAuthUserService(ctrl), "", "", nil)
	cookie := captureStateCookie(t, mux)

	req := httptest.NewRequest(http.MethodGet, "/auth/twitch/callback?code=x&state=wrong", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAuthHandler_Callback_MissingCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	mux := buildAuthMux(t, mock.NewMockAuthUserService(ctrl), "", "", nil)
	cookie := captureStateCookie(t, mux)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/auth/twitch/callback?state=%s", cookie.Value), nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- Callback: upstream failure paths ---

func TestAuthHandler_Callback_TokenExchangeFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	t.Cleanup(tokenSrv.Close)

	mux := buildAuthMux(t, mock.NewMockAuthUserService(ctrl), tokenSrv.URL, "", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, callbackReq(captureStateCookie(t, mux), "testcode"))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAuthHandler_Callback_TwitchAPIFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	usersSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	t.Cleanup(usersSrv.Close)

	mux := buildAuthMux(t, mock.NewMockAuthUserService(ctrl), mockTokenSrv(t).URL, usersSrv.URL, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, callbackReq(captureStateCookie(t, mux), "testcode"))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- Callback: happy paths ---

func TestAuthHandler_Callback_ExistingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	apiToken := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	existing := &model.StreamingUser{UserName: "streamerAlice", ApiToken: apiToken}

	svc := mock.NewMockAuthUserService(ctrl)
	svc.EXPECT().GetUserByTwitchId(gomock.Any(), "twitch-alice").Return(existing, nil)

	mux := buildAuthMux(t, svc, mockTokenSrv(t).URL, mockUsersSrv(t, "twitch-alice", "streamerAlice").URL, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, callbackReq(captureStateCookie(t, mux), "testcode"))

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, rr.Body.String(), "streamerAlice")
	assert.Contains(t, rr.Body.String(), apiToken.String())
}

func TestAuthHandler_Callback_NewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	apiToken := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	created := &model.StreamingUser{UserName: "newStreamer", ApiToken: apiToken}

	svc := mock.NewMockAuthUserService(ctrl)
	svc.EXPECT().GetUserByTwitchId(gomock.Any(), "twitch-new").Return(nil, sql.ErrNoRows)
	svc.EXPECT().CreateUser(gomock.Any(), "newStreamer", "twitch-new", false).Return(created, nil)

	mux := buildAuthMux(t, svc, mockTokenSrv(t).URL, mockUsersSrv(t, "twitch-new", "newStreamer").URL, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, callbackReq(captureStateCookie(t, mux), "testcode"))

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "newStreamer")
}

func TestAuthHandler_Callback_NewAdminUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	apiToken := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	adminUser := &model.StreamingUser{UserName: "adminStreamer", ApiToken: apiToken, IsAdmin: true}

	svc := mock.NewMockAuthUserService(ctrl)
	svc.EXPECT().GetUserByTwitchId(gomock.Any(), "admin-twitch-id").Return(nil, sql.ErrNoRows)
	svc.EXPECT().CreateUser(gomock.Any(), "adminStreamer", "admin-twitch-id", true).Return(adminUser, nil)

	mux := buildAuthMux(t, svc,
		mockTokenSrv(t).URL, mockUsersSrv(t, "admin-twitch-id", "adminStreamer").URL,
		[]string{"admin-twitch-id"},
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, callbackReq(captureStateCookie(t, mux), "testcode"))

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "adminStreamer")
}

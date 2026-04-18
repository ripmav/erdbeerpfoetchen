package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ripmav/erdbeerpfoetchen/internal/middleware"
)

// stubValidator is a test double for middleware.TokenValidator.
type stubValidator struct {
	token uuid.UUID
}

func (s *stubValidator) GetUserApiToken(_ context.Context, _ string) (uuid.UUID, error) {
	return s.token, nil
}

var testToken = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func okHandler(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/collection/{streamer_name}", middleware.Auth(&stubValidator{token: testToken}, okHandler))
	return mux
}

func TestAuth_MissingAuthorizationHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_MalformedAuthorizationHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"no scheme", testToken.String()},
		{"wrong scheme", "Basic " + testToken.String()},
		{"too many parts", "Bearer token extra"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
			req.Header.Set("Authorization", tt.header)
			rr := httptest.NewRecorder()
			newMux().ServeHTTP(rr, req)
			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}

func TestAuth_EmptyStreamerName(t *testing.T) {
	// Invoke handler directly (no mux) so PathValue("streamer_name") returns "".
	handler := middleware.Auth(&stubValidator{token: testToken}, okHandler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+testToken.String())
	rr := httptest.NewRecorder()
	handler(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
	req.Header.Set("Authorization", "Bearer not-a-uuid")
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAuth_ValidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
	req.Header.Set("Authorization", "Bearer "+testToken.String())
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuth_BearerPrefixCaseInsensitive(t *testing.T) {
	for _, prefix := range []string{"bearer", "BEARER", "Bearer"} {
		t.Run(prefix, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
			req.Header.Set("Authorization", prefix+" "+testToken.String())
			rr := httptest.NewRecorder()
			newMux().ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}

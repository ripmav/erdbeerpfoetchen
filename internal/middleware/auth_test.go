package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ripmav/erdbeerpfoetchen/internal/middleware"
)

func okHandler(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/collection/{streamer_name}", middleware.Auth(okHandler))
	return mux
}

func TestAuth_MissingAuthorizationHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuth_MalformedAuthorizationHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"no scheme", "my-super-secret-key"},
		{"wrong scheme", "Basic my-super-secret-key"},
		{"too many parts", "Bearer token extra"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
			req.Header.Set("Authorization", tt.header)
			rr := httptest.NewRecorder()
			newMux().ServeHTTP(rr, req)
			if rr.Code != http.StatusUnauthorized {
				t.Errorf("got %d, want %d", rr.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAuth_EmptyStreamerName(t *testing.T) {
	// Invoke handler directly (no mux) so PathValue("streamer_name") returns "".
	handler := middleware.Auth(okHandler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer my-super-secret-key")
	rr := httptest.NewRecorder()
	handler(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
	req.Header.Set("Authorization", "Bearer my-super-secret-key")
	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAuth_BearerPrefixCaseInsensitive(t *testing.T) {
	for _, prefix := range []string{"bearer", "BEARER", "Bearer"} {
		t.Run(prefix, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/alice", nil)
			req.Header.Set("Authorization", prefix+" my-super-secret-key")
			rr := httptest.NewRecorder()
			newMux().ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
			}
		})
	}
}

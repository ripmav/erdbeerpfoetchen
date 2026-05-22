package handle

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ripmav/erdbeerpfoetchen/internal/database/model"
	"golang.org/x/oauth2"
)

const (
	stateCookieName   = "twitch_oauth_state"
	stateCookieMaxAge = 5 * time.Minute

	twitchAuthURL  = "https://id.twitch.tv/oauth2/authorize"
	twitchTokenURL = "https://id.twitch.tv/oauth2/token"
	twitchUsersURL = "https://api.twitch.tv/helix/users"
)

type AuthUserService interface {
	GetUserByTwitchId(ctx context.Context, twitchID string) (*model.StreamingUser, error)
	CreateUser(ctx context.Context, userName, twitchID string, isAdmin bool) (*model.StreamingUser, error)
}

type AuthHandler struct {
	oauth      *oauth2.Config
	users      AuthUserService
	adminIDs   map[string]struct{}
	usersURL   string
	httpClient *http.Client
	secure     bool
}

func newAuthHandler(cfg *oauth2.Config, adminIDs []string, users AuthUserService) *AuthHandler {
	adminMap := make(map[string]struct{}, len(adminIDs))
	for _, id := range adminIDs {
		adminMap[id] = struct{}{}
	}
	return &AuthHandler{
		oauth:      cfg,
		users:      users,
		adminIDs:   adminMap,
		usersURL:   twitchUsersURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func NewAuthHandler(clientID, clientSecret, redirectURL string, adminIDs []string, users AuthUserService, secure bool) *AuthHandler {
	h := newAuthHandler(&oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  twitchAuthURL,
			TokenURL: twitchTokenURL,
		},
	}, adminIDs, users)
	h.secure = secure
	return h
}

func (h *AuthHandler) Install(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.loginPage)
	mux.HandleFunc("GET /auth/twitch", h.login)
	mux.HandleFunc("GET /auth/twitch/callback", h.callback)
}

func (h *AuthHandler) loginPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := loginTmpl.Execute(w, nil); err != nil {
		slog.ErrorContext(r.Context(), "failed to render login page", "error", err)
	}
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to generate oauth state", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		MaxAge:   int(stateCookieMaxAge / time.Second),
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.Redirect(w, r, h.oauth.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *AuthHandler) callback(w http.ResponseWriter, r *http.Request) {
	if err := h.verifyState(r); err != nil {
		h.renderError(w, r, http.StatusForbidden, "Invalid state — please try again.")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   stateCookieName,
		MaxAge: -1,
		Path:   "/",
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		h.renderError(w, r, http.StatusBadRequest, "Missing authorization code from Twitch.")
		return
	}

	token, err := h.oauth.Exchange(r.Context(), code)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to exchange oauth code", "error", err)
		h.renderError(w, r, http.StatusInternalServerError, "Could not complete Twitch authentication.")
		return
	}

	twitchID, login, err := h.fetchTwitchUser(r.Context(), token.AccessToken)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch twitch user info", "error", err)
		h.renderError(w, r, http.StatusInternalServerError, "Could not retrieve your Twitch account details.")
		return
	}

	u, err := h.getOrCreateUser(r.Context(), twitchID, login)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get or create user", "error", err)
		h.renderError(w, r, http.StatusInternalServerError, "Could not create or load your account.")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := successTmpl.Execute(w, map[string]any{
		"UserName": u.UserName,
		"ApiToken": u.ApiToken,
	}); err != nil {
		slog.ErrorContext(r.Context(), "failed to render success page", "error", err)
	}
}

func (h *AuthHandler) renderError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	if err := errorTmpl.Execute(w, msg); err != nil {
		slog.ErrorContext(r.Context(), "failed to render error page", "error", err)
	}
}

func (h *AuthHandler) verifyState(r *http.Request) error {
	cookie, err := r.Cookie(stateCookieName)
	if err != nil {
		return errors.New("missing state cookie")
	}
	if cookie.Value != r.URL.Query().Get("state") {
		return errors.New("state mismatch")
	}
	return nil
}

type twitchUsersResponse struct {
	Data []struct {
		ID    string `json:"id"`
		Login string `json:"login"`
	} `json:"data"`
}

func (h *AuthHandler) fetchTwitchUser(ctx context.Context, accessToken string) (id, login string, _ error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.usersURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Client-Id", h.oauth.ClientID)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("twitch API returned status %d", resp.StatusCode)
	}

	var result twitchUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	if len(result.Data) == 0 {
		return "", "", errors.New("no user data in Twitch response")
	}

	return result.Data[0].ID, result.Data[0].Login, nil
}

func (h *AuthHandler) getOrCreateUser(ctx context.Context, twitchID, login string) (*model.StreamingUser, error) {
	u, err := h.users.GetUserByTwitchId(ctx, twitchID)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	_, isAdmin := h.adminIDs[twitchID]
	return h.users.CreateUser(ctx, login, twitchID, isAdmin)
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

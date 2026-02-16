package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := extractToken(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		streamerName := r.PathValue("streamer_name")

		if streamerName == "" {
			http.Error(w, "Forbidden: Streamer name is required in URI", http.StatusForbidden)
			return
		}

		if !isValid(r.Context(), token, streamerName) {
			http.Error(w, "Forbidden: Invalid token", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	// Split the header into "Bearer" and "<token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

func isValid(ctx context.Context, token, streamerName string) bool {

	// TODO: Validate the token against a database
	const secretToken = "my-super-secret-key"
	return token == secretToken
}

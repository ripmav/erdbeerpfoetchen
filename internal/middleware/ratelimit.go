package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"
)

const defaultRatePerMinute = 100

// RateLimitProvider returns the configured rate limit (requests per minute) for a streamer.
type RateLimitProvider interface {
	GetUserRateLimit(ctx context.Context, userName string) (int32, error)
}

// RateLimiter implements a per-streamer token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*bucket
	provider RateLimitProvider
}

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// NewRateLimiter creates a RateLimiter that fetches per-streamer rate limits from provider.
// If provider is nil or the lookup fails, the default of 100 requests/minute is used.
func NewRateLimiter(provider RateLimitProvider) *RateLimiter {
	rl := &RateLimiter{
		clients:  make(map[string]*bucket),
		provider: provider,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for key, b := range rl.clients {
			if time.Since(b.lastSeen) > 3*time.Minute {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) allow(key string, rate, burst float64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.clients[key]
	if !ok {
		b = &bucket{tokens: burst, lastSeen: now}
		rl.clients[key] = b
	}

	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * rate
	if b.tokens > burst {
		b.tokens = burst
	}
	b.lastSeen = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// Limit wraps next with per-streamer rate limiting, returning 429 when the limit is exceeded.
// The rate limit is read from the database per streamer; defaults to 100 requests/minute on error.
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streamerName := r.PathValue("streamer_name")

		ratePerMin := int32(defaultRatePerMinute)
		if rl.provider != nil && streamerName != "" {
			if limit, err := rl.provider.GetUserRateLimit(r.Context(), streamerName); err == nil {
				ratePerMin = limit
			}
		}

		rate := float64(ratePerMin) / 60.0
		burst := float64(ratePerMin)

		if !rl.allow(streamerName, rate, burst) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

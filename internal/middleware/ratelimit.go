package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements a per-IP token bucket rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*bucket
	rate    float64 // tokens replenished per second
	burst   float64 // maximum token capacity
}

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// NewRateLimiter creates a RateLimiter allowing rate requests/second with a
// maximum burst of burst requests. Stale client state is cleaned up every minute.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*bucket),
		rate:    rate,
		burst:   float64(burst),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, b := range rl.clients {
			if time.Since(b.lastSeen) > 3*time.Minute {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.clients[ip]
	if !ok {
		b = &bucket{tokens: rl.burst, lastSeen: now}
		rl.clients[ip] = b
	}

	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.lastSeen = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// Limit wraps next with per-IP rate limiting, returning 429 when the limit is exceeded.
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if i := strings.LastIndex(ip, ":"); i != -1 {
			ip = ip[:i]
		}
		if !rl.allow(ip) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

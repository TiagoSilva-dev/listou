package httpx

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter is a small in-process fixed-window limiter. It is enough for a
// single API instance; when we scale horizontally it is the one place that
// moves to Redis (see docs/decisions.md, ADR-0004).
type RateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string]*windowCount
	now    func() time.Time
}

type windowCount struct {
	start time.Time
	count int
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{limit: limit, window: window, hits: map[string]*windowCount{}, now: time.Now}
}

func (l *RateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	wc, ok := l.hits[key]
	if !ok || now.Sub(wc.start) >= l.window {
		if len(l.hits) > 50_000 {
			l.sweep(now)
		}
		l.hits[key] = &windowCount{start: now, count: 1}
		return true
	}
	wc.count++
	return wc.count <= l.limit
}

func (l *RateLimiter) sweep(now time.Time) {
	for k, wc := range l.hits {
		if now.Sub(wc.start) >= l.window {
			delete(l.hits, k)
		}
	}
}

// Limit applies the limiter per client IP and route bucket.
func (l *RateLimiter) Limit(bucket string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow(bucket + "|" + ClientIP(r)) {
				w.Header().Set("Retry-After", "60")
				Fail(w, r, ErrRateLimited)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClientIP uses the first X-Forwarded-For hop (set by our own proxy/BFF)
// and falls back to the socket address. Only used for rate limiting.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first, _, _ := strings.Cut(xff, ",")
		return strings.TrimSpace(first)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

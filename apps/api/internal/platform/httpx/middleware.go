package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/listou/listou/apps/api/internal/platform/logger"
)

type requestIDKey struct{}

const RequestIDHeader = "X-Request-Id"

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func newRequestID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RequestID accepts a well-formed inbound id (e.g. from the BFF) or mints one,
// and attaches a request-scoped logger carrying it.
func RequestID(base *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(RequestIDHeader)
			if !validRequestID.MatchString(id) {
				id = newRequestID()
			}
			w.Header().Set(RequestIDHeader, id)
			ctx := context.WithValue(r.Context(), requestIDKey{}, id)
			ctx = logger.WithContext(ctx, base.With("request_id", id))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// UserIDFunc lets the auth module enrich access logs without httpx importing it.
type UserIDFunc func(context.Context) string

// AccessLog logs one structured line per request. It never logs bodies,
// query strings or headers, which may carry tokens or PII.
func AccessLog(userID UserIDFunc) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			holder := &userHolder{}
			ctx := context.WithValue(r.Context(), userHolderKey{}, holder)
			next.ServeHTTP(rec, r.WithContext(ctx))
			attrs := []any{
				"method", r.Method,
				"endpoint", routePattern(r),
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			}
			if holder.id != "" {
				attrs = append(attrs, "user_id", holder.id)
			}
			l := logger.From(ctx)
			if rec.status >= 500 {
				l.Error("http_request", attrs...)
			} else {
				l.Info("http_request", attrs...)
			}
		})
	}
}

type userHolderKey struct{}
type userHolder struct{ id string }

// SetLogUserID records the authenticated user id for the access log line.
func SetLogUserID(ctx context.Context, id string) {
	if h, ok := ctx.Value(userHolderKey{}).(*userHolder); ok {
		h.id = id
	}
}

func routePattern(r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}
	return r.URL.Path
}

func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.From(r.Context()).Error("panic recovered", "panic", fmt.Sprint(rec), "stack", string(debug.Stack()))
					Fail(w, r, ErrInternal)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func SecureHeaders(production bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			if production {
				h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

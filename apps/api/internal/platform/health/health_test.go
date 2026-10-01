package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func TestProbes(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		db     error
		status int
	}{
		{"live ok", "/health", errors.New("db down"), http.StatusOK},
		{"ready ok", "/ready", nil, http.StatusOK},
		{"ready down", "/ready", errors.New("db down"), http.StatusServiceUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			NewHandler(fakePinger{tc.db}, "test").Register(mux)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != tc.status {
				t.Fatalf("got %d want %d", rec.Code, tc.status)
			}
		})
	}
}

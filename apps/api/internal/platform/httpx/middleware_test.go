package httpx

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func discard() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestRequestIDAndErrorEnvelope(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Fail(w, r, NotFound("LIST_NOT_FOUND", "Lista não encontrada."))
	}), RequestID(discard()), Recover())

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(RequestIDHeader, "abcDEF123456")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
	var body struct {
		Error struct{ Code, Message, RequestID string }
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "LIST_NOT_FOUND" || body.Error.RequestID != "abcDEF123456" {
		t.Fatalf("unexpected body %+v", body)
	}
}

func TestRequestIDRejectsMalformed(t *testing.T) {
	h := RequestID(discard())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "bad id\nwith newline")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get(RequestIDHeader); len(got) != 24 {
		t.Fatalf("expected generated id, got %q", got)
	}
}

func TestRecoverHidesPanic(t *testing.T) {
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("secret detail") }), RequestID(discard()), Recover())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("panic leaked: %d %s", rec.Code, rec.Body.String())
	}
}

func TestSameOrigin(t *testing.T) {
	h := SameOrigin("http://localhost:3000")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for origin, want := range map[string]int{
		"":                      http.StatusNoContent,
		"http://localhost:3000": http.StatusNoContent,
		"https://evil.example":  http.StatusForbidden,
	} {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Errorf("origin %q: got %d want %d", origin, rec.Code, want)
		}
	}
}

func TestRateLimiter(t *testing.T) {
	l := NewRateLimiter(2, time.Minute)
	now := time.Unix(0, 0)
	l.now = func() time.Time { return now }
	if !l.Allow("k") || !l.Allow("k") || l.Allow("k") {
		t.Fatal("expected third hit to be limited")
	}
	now = now.Add(time.Minute)
	if !l.Allow("k") {
		t.Fatal("expected window reset")
	}
}

func TestDecodeRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":1,"b":2}`))
	var dst struct {
		A int `json:"a"`
	}
	if err := Decode(req, &dst); AsError(err).Code != "INVALID_JSON" {
		t.Fatalf("expected INVALID_JSON, got %v", err)
	}
}

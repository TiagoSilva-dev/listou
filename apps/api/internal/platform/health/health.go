// Package health exposes liveness (/health) and readiness (/ready) probes.
package health

import (
	"context"
	"net/http"
	"time"

	"github.com/listou/listou/apps/api/internal/platform/httpx"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	db      Pinger
	version string
}

func NewHandler(db Pinger, version string) *Handler {
	return &Handler{db: db, version: version}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.live)
	mux.HandleFunc("GET /ready", h.ready)
}

func (h *Handler) live(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok", "version": h.version})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		httpx.JSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "unavailable",
			"checks": map[string]string{"database": "down"},
		})
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"status": "ready",
		"checks": map[string]string{"database": "up"},
	})
}

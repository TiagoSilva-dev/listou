// Package app wires the modules of the modular monolith into one HTTP handler.
package app

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/listou/listou/apps/api/internal/platform/config"
	"github.com/listou/listou/apps/api/internal/platform/flags"
	"github.com/listou/listou/apps/api/internal/platform/health"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
)

type App struct {
	cfg     config.Config
	pool    *pgxpool.Pool
	log     *slog.Logger
	flags   flags.Set
	version string
}

func New(cfg config.Config, pool *pgxpool.Pool, log *slog.Logger, version string) *App {
	return &App{cfg: cfg, pool: pool, log: log, flags: flags.Parse(cfg.FeatureFlags), version: version}
}

func (a *App) Handler() http.Handler {
	root := http.NewServeMux()
	health.NewHandler(a.pool, a.version).Register(root)

	api := http.NewServeMux()
	api.HandleFunc("GET /api/v1/flags", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]any{"flags": a.flags.All()})
	})
	root.Handle("/api/v1/", api)
	root.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpx.Fail(w, r, httpx.NotFound("ROUTE_NOT_FOUND", "Rota não encontrada."))
	})

	return httpx.Chain(root,
		httpx.RequestID(a.log),
		httpx.AccessLog(nil),
		httpx.Recover(),
		httpx.SecureHeaders(a.cfg.IsProduction()),
		httpx.SameOrigin(a.cfg.PublicWebURL),
	)
}

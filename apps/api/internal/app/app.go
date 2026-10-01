// Package app wires the modules of the modular monolith into one HTTP handler.
package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/listou/listou/apps/api/internal/affiliate"
	"github.com/listou/listou/apps/api/internal/affiliate/mock"
	"github.com/listou/listou/apps/api/internal/analytics"
	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/catalog"
	"github.com/listou/listou/apps/api/internal/dashboard"
	"github.com/listou/listou/apps/api/internal/decision"
	"github.com/listou/listou/apps/api/internal/events"
	"github.com/listou/listou/apps/api/internal/listbuilder"
	"github.com/listou/listou/apps/api/internal/lists"
	"github.com/listou/listou/apps/api/internal/platform/config"
	"github.com/listou/listou/apps/api/internal/platform/flags"
	"github.com/listou/listou/apps/api/internal/platform/health"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/publiclists"
	"github.com/listou/listou/apps/api/internal/recommendations"
	"github.com/listou/listou/apps/api/internal/reservations"
)

type App struct {
	cfg          config.Config
	pool         *pgxpool.Pool
	log          *slog.Logger
	flags        flags.Set
	version      string
	reservations *reservations.Service
}

func New(cfg config.Config, pool *pgxpool.Pool, log *slog.Logger, version string) *App {
	return &App{cfg: cfg, pool: pool, log: log, flags: flags.Parse(cfg.FeatureFlags), version: version}
}

// RunBackground starts periodic jobs and stops when ctx is cancelled.
func (a *App) RunBackground(ctx context.Context) {
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if n, err := a.reservations.ExpireDue(ctx); err != nil {
					a.log.Error("expire reservations", "error", err.Error())
				} else if n > 0 {
					a.log.Info("reservations expired", "units", n)
				}
			}
		}
	}()
}

func (a *App) Handler() http.Handler {
	secure := a.cfg.CookieSecure

	recorder := analytics.NewRecorder(a.pool)
	registry := affiliate.NewRegistry(mock.New(a.cfg.PublicWebURL))
	recs := recommendations.RulesProvider{}
	decider := decision.RuleEngine{}

	authSvc := auth.NewService(a.pool, a.cfg.SessionTTL, recorder)
	authH := auth.NewHandler(authSvc, auth.CookieConfig{Secure: secure, Domain: a.cfg.CookieDomain})

	catalogSvc := catalog.NewService(a.pool, registry, recorder)
	listsSvc := lists.NewService(a.pool, catalogSvc, decider, recorder)
	eventsSvc := events.NewService(a.pool, listsSvc, recs, recorder)
	a.reservations = reservations.NewService(a.pool, a.cfg.ReservationTTL, recorder)
	publicSvc := publiclists.NewService(a.pool, listsSvc)
	dashSvc := dashboard.NewService(a.pool, eventsSvc)

	authLimit := httpx.NewRateLimiter(20, time.Minute)
	guestLimit := httpx.NewRateLimiter(30, time.Minute)
	trackLimit := httpx.NewRateLimiter(120, time.Minute)
	goLimit := httpx.NewRateLimiter(120, time.Minute)

	root := http.NewServeMux()
	health.NewHandler(a.pool, a.version).Register(root)
	root.HandleFunc("GET /api/v1/flags", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]any{"flags": a.flags.All()})
	})

	authH.Register(root, authLimit.Limit("auth"))
	events.NewHandler(eventsSvc).Register(root, authH.RequireFunc)
	lists.NewHandler(listsSvc).Register(root, authH.RequireFunc)
	catalog.NewHandler(catalogSvc).Register(root, authH.RequireFunc)
	recommendations.NewHandler(recs).Register(root)
	builderSvc := listbuilder.NewService(eventsSvc, listsSvc, recommendations.NewBuilder(recs), recorder, a.flags.Enabled(flags.AIListBuilder))
	listbuilder.NewHandler(builderSvc).Register(root, authH.RequireFunc)
	publiclists.NewHandler(publicSvc, authH.Optional).Register(root)
	reservations.NewHandler(a.reservations, secure).Register(root, guestLimit.Limit("guest"), authH.RequireFunc)
	analytics.NewHandler(recorder, a.pool, secure).Register(root, trackLimit.Limit("track"))
	dashboard.NewHandler(dashSvc).Register(root, authH.RequireFunc)
	affiliate.NewRedirector(a.pool, registry, recorder, secure).Register(root, goLimit.Limit("go"))

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

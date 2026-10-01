// Package admin is the back-office: endpoints restricted to ADMIN users, used
// to maintain the curated product catalog (affiliate links we generate).
package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/listou/listou/apps/api/internal/affiliate/curated"
	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/platform/audit"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/logger"
)

var errProductNotFound = httpx.NotFound("CURATED_PRODUCT_NOT_FOUND", "Produto não encontrado.")

// Catalog is the storage the panel manages (implemented by curated.Repository).
type Catalog interface {
	List(ctx context.Context, onlyActive bool) ([]curated.Product, error)
	Get(ctx context.Context, id uuid.UUID) (curated.Product, error)
	Merchants(ctx context.Context) ([]curated.Merchant, error)
	Create(ctx context.Context, in curated.Input) (uuid.UUID, error)
	Update(ctx context.Context, id uuid.UUID, in curated.Input) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
}

type Service struct {
	db      database.DBTX
	catalog Catalog
}

func NewService(pool *pgxpool.Pool, c Catalog) *Service { return &Service{db: pool, catalog: c} }

func (s *Service) List(ctx context.Context) ([]curated.Product, []curated.Merchant, error) {
	products, err := s.catalog.List(ctx, false)
	if err != nil {
		return nil, nil, err
	}
	merchants, err := s.catalog.Merchants(ctx)
	if err != nil {
		return nil, nil, err
	}
	if products == nil {
		products = []curated.Product{}
	}
	return products, merchants, nil
}

func (s *Service) allowed(ctx context.Context) (map[string]bool, error) {
	ms, err := s.catalog.Merchants(ctx)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(ms))
	for _, m := range ms {
		set[m.Code] = true
	}
	return set, nil
}

func (s *Service) Create(ctx context.Context, actor auth.User, in curated.Input) (curated.Product, error) {
	allowed, err := s.allowed(ctx)
	if err != nil {
		return curated.Product{}, err
	}
	if err := in.Validate(allowed); err != nil {
		return curated.Product{}, err
	}
	id, err := s.catalog.Create(ctx, in)
	if err != nil {
		return curated.Product{}, err
	}
	s.audit(ctx, actor, "CURATED_PRODUCT_CREATED", id, in)
	return s.catalog.Get(ctx, id)
}

func (s *Service) Update(ctx context.Context, actor auth.User, id uuid.UUID, in curated.Input) (curated.Product, error) {
	allowed, err := s.allowed(ctx)
	if err != nil {
		return curated.Product{}, err
	}
	if err := in.Validate(allowed); err != nil {
		return curated.Product{}, err
	}
	found, err := s.catalog.Update(ctx, id, in)
	if err != nil {
		return curated.Product{}, err
	}
	if !found {
		return curated.Product{}, errProductNotFound
	}
	s.audit(ctx, actor, "CURATED_PRODUCT_UPDATED", id, in)
	return s.catalog.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, actor auth.User, id uuid.UUID) error {
	found, err := s.catalog.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !found {
		return errProductNotFound
	}
	s.audit(ctx, actor, "CURATED_PRODUCT_DELETED", id, curated.Input{})
	return nil
}

// audit is best-effort: the change already happened, so a failed audit row is
// logged instead of turned into an error.
func (s *Service) audit(ctx context.Context, actor auth.User, action string, id uuid.UUID, in curated.Input) {
	meta := map[string]any{}
	if in.Title != "" {
		meta["title"] = in.Title
		meta["active"] = in.Active
		meta["merchants"] = len(in.Offers)
	}
	if err := audit.Log(ctx, s.db, &actor.ID, action, "curated_product", id, meta); err != nil {
		logger.From(ctx).Warn("audit not recorded", "action", action, "error", err.Error())
	}
}

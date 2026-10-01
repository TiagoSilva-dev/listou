package catalog

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/listou/listou/apps/api/internal/affiliate"
	"github.com/listou/listou/apps/api/internal/analytics"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/logger"
)

var ErrProductNotFound = httpx.NotFound("PRODUCT_NOT_FOUND", "Produto não encontrado.")

type Service struct {
	pool     *pgxpool.Pool
	repo     Repository
	registry *affiliate.Registry
	matcher  MatchingService
	tracker  *analytics.Recorder
	now      func() time.Time
}

func NewService(pool *pgxpool.Pool, registry *affiliate.Registry, tracker *analytics.Recorder) *Service {
	return &Service{pool: pool, registry: registry, tracker: tracker, now: time.Now}
}

type SearchParams struct {
	Query    string
	Sort     string // relevance | price_asc | price_desc
	Category string // optional hint used for relevance
	Limit    int
	UserID   *uuid.UUID
}

func (s *Service) Search(ctx context.Context, p SearchParams) ([]SearchResult, error) {
	p.Query = strings.TrimSpace(p.Query)
	if len(p.Query) < 2 || len(p.Query) > 100 {
		return nil, httpx.Validation(map[string]string{"q": "Digite de 2 a 100 caracteres."})
	}
	merchants, err := s.repo.EnabledMerchants(ctx, s.pool)
	if err != nil {
		return nil, err
	}
	byAdapter := map[string]map[string]Merchant{}
	for _, m := range merchants {
		if byAdapter[m.Adapter] == nil {
			byAdapter[m.Adapter] = map[string]Merchant{}
		}
		byAdapter[m.Adapter][m.Code] = Merchant{Code: m.Code, Name: m.Name}
	}

	var results []SearchResult
	for adapter, allowed := range byAdapter {
		provider, ok := s.registry.Get(adapter)
		if !ok {
			continue
		}
		products, err := provider.SearchProducts(ctx, affiliate.SearchQuery{Text: p.Query, Limit: p.Limit})
		if err != nil {
			// One failing provider must not break discovery for the others.
			logger.From(ctx).Warn("provider search failed", "adapter", adapter, "error", err.Error())
			continue
		}
		for _, pd := range products {
			if r, ok := toResult(adapter, pd, allowed); ok {
				results = append(results, r)
			}
		}
	}

	for i := range results {
		r := &results[i]
		score, verdict := s.matcher.Score(p.Query, p.Category, Candidate{Title: r.Title, Brand: deref(r.Brand), Category: deref(r.Category)})
		r.Match = &Ranked{Index: i, Score: score, Verdict: verdict}
	}
	sortResults(results, p.Sort)
	if p.Limit > 0 && len(results) > p.Limit {
		results = results[:p.Limit]
	}
	s.tracker.Track(ctx, analytics.Event{Name: analytics.ProductSearched, UserID: p.UserID,
		Props: map[string]any{"q": p.Query, "results": float64(len(results))}})
	return results, nil
}

func toResult(adapter string, pd affiliate.ProductData, allowed map[string]Merchant) (SearchResult, bool) {
	r := SearchResult{
		ProviderCode: adapter, ExternalID: pd.ExternalID, Title: pd.Title, Brand: pd.Brand,
		ImageURL: pd.ImageURL, Emoji: pd.Emoji, Category: pd.Category, Rating: pd.Rating, Demo: pd.Demo,
	}
	for _, o := range pd.Offers {
		m, ok := allowed[o.MerchantCode]
		if !ok {
			continue
		}
		r.Offers = append(r.Offers, SearchOffer{Merchant: m, PriceCents: o.PriceCents, Availability: string(o.Availability)})
		if o.PriceCents != nil && (r.LowestPriceCents == nil || *o.PriceCents < *r.LowestPriceCents) {
			price := *o.PriceCents
			r.LowestPriceCents = &price
		}
	}
	return r, len(r.Offers) > 0
}

func sortResults(rs []SearchResult, mode string) {
	price := func(r SearchResult) int64 {
		if r.LowestPriceCents == nil {
			return 1 << 62
		}
		return *r.LowestPriceCents
	}
	switch mode {
	case "price_asc":
		sort.SliceStable(rs, func(i, j int) bool { return price(rs[i]) < price(rs[j]) })
	case "price_desc":
		sort.SliceStable(rs, func(i, j int) bool { return price(rs[i]) > price(rs[j]) })
	default:
		sort.SliceStable(rs, func(i, j int) bool { return rs[i].Match.Score > rs[j].Match.Score })
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Import fetches a product from its provider and stores it with its offers.
func (s *Service) Import(ctx context.Context, adapter, externalID string) (Product, []Offer, error) {
	provider, ok := s.registry.Get(adapter)
	if !ok {
		return Product{}, nil, httpx.BadRequest("UNKNOWN_PROVIDER", "Loja não disponível.")
	}
	pd, err := provider.GetProduct(ctx, externalID)
	if err != nil {
		if errors.Is(err, affiliate.ErrNotFound) {
			return Product{}, nil, ErrProductNotFound
		}
		return Product{}, nil, fmt.Errorf("import: %w", err)
	}
	if strings.TrimSpace(pd.Title) == "" {
		return Product{}, nil, ErrLinkUnreadable
	}
	return s.store(ctx, adapter, pd)
}

func (s *Service) store(ctx context.Context, adapter string, pd affiliate.ProductData) (Product, []Offer, error) {
	merchants, err := s.repo.EnabledMerchants(ctx, s.pool)
	if err != nil {
		return Product{}, nil, err
	}
	ids := map[string]uuid.UUID{}
	for _, m := range merchants {
		if m.Adapter == adapter {
			ids[m.Code] = m.ID
		}
	}
	var productID uuid.UUID
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		productID, err = s.repo.UpsertProduct(ctx, tx, pd, ids, s.now())
		return err
	})
	if err != nil {
		return Product{}, nil, err
	}
	return s.Get(ctx, productID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Product, []Offer, error) {
	p, err := s.repo.Product(ctx, s.pool, id)
	if err != nil {
		if database.IsNoRows(err) {
			return Product{}, nil, ErrProductNotFound
		}
		return Product{}, nil, err
	}
	offers, err := s.repo.OffersForProducts(ctx, s.pool, []uuid.UUID{id})
	if err != nil {
		return Product{}, nil, err
	}
	return p, nonNil(offers[id]), nil
}

// Hydrate loads products and offers for many items at once (no N+1).
func (s *Service) Hydrate(ctx context.Context, db database.DBTX, productIDs []uuid.UUID) (map[uuid.UUID]Product, map[uuid.UUID][]Offer, error) {
	products, err := s.repo.Products(ctx, db, productIDs)
	if err != nil {
		return nil, nil, err
	}
	offers, err := s.repo.OffersForProducts(ctx, db, productIDs)
	if err != nil {
		return nil, nil, err
	}
	return products, offers, nil
}

func nonNil(o []Offer) []Offer {
	if o == nil {
		return []Offer{}
	}
	return o
}

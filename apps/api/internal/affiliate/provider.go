// Package affiliate defines the marketplace-agnostic Provider contract.
// Marketplace-specific code lives only in subpackages (mock, and future
// amazon/mercadolivre/shopee adapters). See docs/affiliate-providers.md.
package affiliate

import (
	"context"
	"errors"
)

type Availability string

const (
	InStock    Availability = "IN_STOCK"
	OutOfStock Availability = "OUT_OF_STOCK"
	Unknown    Availability = "UNKNOWN"
)

type SearchQuery struct {
	Text  string
	Limit int
}

// OfferData is a merchant listing as returned by a provider.
type OfferData struct {
	MerchantCode       string
	ExternalID         string
	Title              string
	PriceCents         *int64
	OriginalPriceCents *int64
	Currency           string
	Availability       Availability
	ProductURL         string
	ImageURL           *string
}

// ProductData is a provider's view of a product and its offers.
type ProductData struct {
	ExternalID  string
	Title       string
	Brand       *string
	Description *string
	GTIN        *string
	ImageURL    *string
	Category    *string
	Emoji       *string
	// Rating is only populated when the source's terms allow displaying it.
	Rating *float64
	// Demo marks fictitious data that must be labeled as such in the UI.
	Demo   bool
	Offers []OfferData
}

// ClickContext carries non-personal attribution context for outbound links.
type ClickContext struct {
	ClickID     string
	EventSlug   string
	UTMSource   string
	UTMMedium   string
	UTMCampaign string
}

type Provider interface {
	// Adapter is the stable adapter code stored in affiliate_providers.adapter.
	Adapter() string
	SearchProducts(ctx context.Context, q SearchQuery) ([]ProductData, error)
	GetProduct(ctx context.Context, externalID string) (ProductData, error)
	GetOffers(ctx context.Context, externalID string) ([]OfferData, error)
	// BuildAffiliateURL turns a stored offer URL into the outbound URL the
	// partner program allows. It must return an absolute http(s) URL.
	BuildAffiliateURL(ctx context.Context, merchantCode, productURL string, click ClickContext) (string, error)
}

var ErrNotFound = errors.New("affiliate: product not found")

// Registry maps adapter codes to implementations.
type Registry struct{ providers map[string]Provider }

func NewRegistry(ps ...Provider) *Registry {
	r := &Registry{providers: map[string]Provider{}}
	for _, p := range ps {
		r.providers[p.Adapter()] = p
	}
	return r
}

func (r *Registry) Get(adapter string) (Provider, bool) {
	p, ok := r.providers[adapter]
	return p, ok
}

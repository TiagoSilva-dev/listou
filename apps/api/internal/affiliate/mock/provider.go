// Package mock is the MockAffiliateProvider: a fictitious catalog used for
// development, demos and tests. Its offer URLs point at the web app's demo
// store page — never at real marketplace URLs.
package mock

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/listou/listou/apps/api/internal/affiliate"
	"github.com/listou/listou/apps/api/internal/catalog/textnorm"
)

const Adapter = "MOCK"

type Provider struct {
	webURL string
}

func New(publicWebURL string) *Provider {
	return &Provider{webURL: strings.TrimRight(publicWebURL, "/")}
}

func (p *Provider) Adapter() string { return Adapter }

func (p *Provider) SearchProducts(_ context.Context, q affiliate.SearchQuery) ([]affiliate.ProductData, error) {
	terms := textnorm.Tokens(q.Text)
	type scored struct {
		f     fixture
		score int
	}
	var hits []scored
	for _, f := range fixtures {
		hay := textnorm.Normalize(f.title + " " + f.keywords + " " + f.category)
		score := 0
		for _, t := range terms {
			if strings.Contains(hay, t) {
				score++
			}
		}
		if len(terms) == 0 || score > 0 {
			hits = append(hits, scored{f, score})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	limit := q.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	out := make([]affiliate.ProductData, 0, min(limit, len(hits)))
	for i := 0; i < len(hits) && i < limit; i++ {
		out = append(out, p.product(hits[i].f))
	}
	return out, nil
}

func (p *Provider) GetProduct(_ context.Context, externalID string) (affiliate.ProductData, error) {
	for _, f := range fixtures {
		if f.id == externalID {
			return p.product(f), nil
		}
	}
	return affiliate.ProductData{}, affiliate.ErrNotFound
}

func (p *Provider) GetOffers(ctx context.Context, externalID string) ([]affiliate.OfferData, error) {
	prod, err := p.GetProduct(ctx, externalID)
	if err != nil {
		return nil, err
	}
	return prod.Offers, nil
}

// BuildAffiliateURL appends clearly fake attribution parameters so the full
// redirect flow can be exercised end to end.
func (p *Provider) BuildAffiliateURL(_ context.Context, merchantCode, productURL string, click affiliate.ClickContext) (string, error) {
	u, err := url.Parse(productURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("mock: invalid product url %q", productURL)
	}
	qs := u.Query()
	qs.Set("demo_tag", "listou-"+strings.ToLower(merchantCode))
	if click.ClickID != "" {
		qs.Set("click", click.ClickID)
	}
	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func (p *Provider) product(f fixture) affiliate.ProductData {
	brand, cat, emoji := f.brand, f.category, f.emoji
	desc := "Produto de demonstração — dados fictícios."
	pd := affiliate.ProductData{
		ExternalID: f.id, Title: f.title, Brand: &brand, Category: &cat, Emoji: &emoji,
		Description: &desc, Demo: true,
	}
	codes := make([]string, 0, len(f.prices))
	for code := range f.prices {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	for _, code := range codes {
		price := f.prices[code]
		pd.Offers = append(pd.Offers, affiliate.OfferData{
			MerchantCode: code,
			ExternalID:   f.id + "-" + code,
			Title:        f.title,
			PriceCents:   &price,
			Currency:     "BRL",
			Availability: affiliate.InStock,
			ProductURL:   fmt.Sprintf("%s/demo/loja/%s/%s", p.webURL, strings.ToLower(code), f.id),
		})
	}
	return pd
}

// Package link is the LinkProvider: it turns a product URL pasted by the user
// into a wish. It reads only what the page publicly declares for link
// previews (Open Graph / schema.org), never prices or availability, and it
// never talks to a marketplace API. The outbound URL is the pasted one until a
// partner program's official link format is documented in docs/integrations.
package link

import (
	"context"
	"net/url"
	"strings"
	"unicode"

	"github.com/listou/listou/apps/api/internal/affiliate"
)

const (
	Adapter      = "LINK"
	MerchantCode = "LINK"
)

type Provider struct{ f *fetcher }

func New() *Provider { return &Provider{f: newFetcher(false)} }

func (p *Provider) Adapter() string { return Adapter }

// SearchProducts: a pasted-link source has nothing to search.
func (p *Provider) SearchProducts(context.Context, affiliate.SearchQuery) ([]affiliate.ProductData, error) {
	return nil, nil
}

// GetProduct treats externalID as the product URL. It fails with
// affiliate.ErrInvalidURL for URLs we refuse to open. A page that cannot be
// read (bot protection, timeouts, no metadata) is not an error: the result
// has an empty Title and the caller decides how to ask the user.
func (p *Provider) GetProduct(ctx context.Context, externalID string) (affiliate.ProductData, error) {
	u, err := Normalize(externalID)
	if err != nil {
		return affiliate.ProductData{}, affiliate.ErrInvalidURL
	}
	// A literal private/loopback IP never reaches the dialer's check with a
	// friendly error, so refuse it up front too.
	if addr, perr := parseHostIP(u.Hostname()); perr == nil && !ipAllowed(addr) {
		return affiliate.ProductData{}, affiliate.ErrInvalidURL
	}
	if strings.EqualFold(u.Hostname(), "localhost") {
		return affiliate.ProductData{}, affiliate.ErrInvalidURL
	}

	var m meta
	pg, err := p.f.get(ctx, u)
	if err == nil {
		m = parseMeta(pg.body, pg.finalURL)
	}
	// The pasted URL stays the outbound one even when it redirects: short
	// links (meli.la, amzn.to) carry their own attribution and stay valid
	// longer than the expanded page with its per-click tokens.
	finalURL := u
	if isGenericTitle(m.Title, u, m.SiteName) {
		m.Title = ""
	}
	return build(finalURL, m), nil
}

// isGenericTitle spots pages (home, captcha, error) whose title is just the
// store's name, which says nothing about the product.
func isGenericTitle(title string, u *url.URL, siteName string) bool {
	alnum := func(s string) string {
		var b strings.Builder
		for _, r := range strings.ToLower(s) {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				b.WriteRune(r)
			}
		}
		return b.String()
	}
	t := alnum(title)
	if t == "" {
		return false
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	label, _, _ := strings.Cut(host, ".")
	return t == alnum(host) || t == alnum(label) || (siteName != "" && t == alnum(siteName))
}

func build(u *url.URL, m meta) affiliate.ProductData {
	canonical := u.String()
	store := m.SiteName
	if store == "" {
		store = strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	}
	pd := affiliate.ProductData{ExternalID: canonical, Title: m.Title}
	if m.Brand != "" {
		pd.Brand = &m.Brand
	}
	if m.Description != "" {
		pd.Description = &m.Description
	}
	if m.Image != "" {
		pd.ImageURL = &m.Image
	}
	pd.Offers = []affiliate.OfferData{{
		MerchantCode: MerchantCode,
		ExternalID:   canonical,
		Title:        m.Title,
		Currency:     "BRL",
		Availability: affiliate.Unknown,
		ProductURL:   canonical,
		ImageURL:     pd.ImageURL,
		StoreName:    &store,
	}}
	return pd
}

func (p *Provider) GetOffers(ctx context.Context, externalID string) ([]affiliate.OfferData, error) {
	pd, err := p.GetProduct(ctx, externalID)
	if err != nil {
		return nil, err
	}
	return pd.Offers, nil
}

// BuildAffiliateURL returns the pasted URL untouched: no partner program is
// integrated here, so nothing is appended to it.
func (p *Provider) BuildAffiliateURL(_ context.Context, _, productURL string, _ affiliate.ClickContext) (string, error) {
	u, err := Normalize(productURL)
	if err != nil {
		return "", affiliate.ErrInvalidURL
	}
	return u.String(), nil
}

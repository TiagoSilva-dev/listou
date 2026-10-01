package curated

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/listou/listou/apps/api/internal/affiliate"
)

type fakeSource []Product

func (f fakeSource) List(_ context.Context, onlyActive bool) ([]Product, error) {
	var out []Product
	for _, p := range f {
		if !onlyActive || p.Active {
			out = append(out, p)
		}
	}
	return out, nil
}

func sample() fakeSource {
	return fakeSource{
		{ExternalID: "ML-AF", Title: "Fritadeira Air Fryer 11L", Brand: "Philco", Category: "Cozinha", Emoji: "🍳",
			Keywords: "air fryer airfryer fritadeira", Active: true,
			Offers: []Offer{{Merchant: "MERCADO_LIVRE", URL: "https://meli.la/2Ne9wWJ"}}},
		{ExternalID: "ML-OFF", Title: "Air fryer desativada", Keywords: "air fryer", Active: false,
			Offers: []Offer{{Merchant: "MERCADO_LIVRE", URL: "https://meli.la/off"}}},
	}
}

func TestSearchAndOffers(t *testing.T) {
	p := New(sample())
	res, err := p.SearchProducts(context.Background(), affiliate.SearchQuery{Text: "air fryer"})
	if err != nil || len(res) != 1 || res[0].ExternalID != "ML-AF" {
		t.Fatalf("unexpected search (inactive must be hidden): %v %+v", err, res)
	}
	o := res[0].Offers[0]
	if o.PriceCents != nil || o.Availability != affiliate.Unknown || o.ProductURL != "https://meli.la/2Ne9wWJ" || res[0].Demo {
		t.Fatalf("offer must carry the affiliate link and no price: %+v", o)
	}
	if res, _ := p.SearchProducts(context.Background(), affiliate.SearchQuery{Text: "geladeira"}); len(res) != 0 {
		t.Fatalf("unexpected hits: %+v", res)
	}
}

func TestGetProductServesInactiveButNotUnknown(t *testing.T) {
	p := New(sample())
	if _, err := p.GetProduct(context.Background(), "ML-OFF"); err != nil {
		t.Fatalf("inactive products must stay resolvable for existing items: %v", err)
	}
	if _, err := p.GetProduct(context.Background(), "nope"); !errors.Is(err, affiliate.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestBuildAffiliateURLIsUntouched(t *testing.T) {
	p := New(sample())
	got, err := p.BuildAffiliateURL(context.Background(), "MERCADO_LIVRE", "https://meli.la/abc", affiliate.ClickContext{ClickID: "x", UTMSource: "s"})
	if err != nil || got != "https://meli.la/abc" {
		t.Fatalf("got %q %v", got, err)
	}
	if _, err := p.BuildAffiliateURL(context.Background(), "X", "javascript:alert(1)", affiliate.ClickContext{}); err == nil {
		t.Fatal("must reject non-http urls")
	}
}

func TestInputValidate(t *testing.T) {
	allowed := map[string]bool{"MERCADO_LIVRE": true}
	ok := Input{Title: "  Faqueiro 24 peças ", Category: "Cozinha", ImageURL: "https://img.example/a.webp",
		Offers: []Offer{{Merchant: "MERCADO_LIVRE", URL: " https://meli.la/abc "}}}
	if err := ok.Validate(allowed); err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}
	if ok.Title != "Faqueiro 24 peças" || ok.Offers[0].URL != "https://meli.la/abc" {
		t.Fatalf("input not trimmed: %+v", ok)
	}
	for name, in := range map[string]Input{
		"short title":   {Title: "ab", Offers: ok.Offers},
		"no offers":     {Title: "Produto bom"},
		"bad url":       {Title: "Produto bom", Offers: []Offer{{Merchant: "MERCADO_LIVRE", URL: "javascript:alert(1)"}}},
		"unknown store": {Title: "Produto bom", Offers: []Offer{{Merchant: "AMAZON", URL: "https://amzn.to/x"}}},
		"duplicate":     {Title: "Produto bom", Offers: []Offer{{Merchant: "MERCADO_LIVRE", URL: "https://a.com"}, {Merchant: "MERCADO_LIVRE", URL: "https://b.com"}}},
		"bad image":     {Title: "Produto bom", ImageURL: "data:image/png;base64,xx", Offers: ok.Offers},
		"huge keywords": {Title: "Produto bom", Keywords: strings.Repeat("a", 501), Offers: ok.Offers},
	} {
		if err := in.Validate(allowed); err == nil {
			t.Errorf("%s should fail", name)
		}
	}
}

func TestEmojiFor(t *testing.T) {
	if EmojiFor(" Cozinha ") != "🍳" || EmojiFor("Qualquer coisa") != "🎁" {
		t.Fatal("unexpected emoji mapping")
	}
}

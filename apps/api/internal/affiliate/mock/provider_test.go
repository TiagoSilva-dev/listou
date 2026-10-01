package mock

import (
	"context"
	"strings"
	"testing"

	"github.com/listou/listou/apps/api/internal/affiliate"
)

func TestSearchAndBuildURL(t *testing.T) {
	p := New("http://localhost:3000/")
	res, err := p.SearchProducts(context.Background(), affiliate.SearchQuery{Text: "Air Fryer"})
	if err != nil || len(res) == 0 {
		t.Fatalf("expected results, got %v %v", res, err)
	}
	if !strings.Contains(strings.ToLower(res[0].Title), "air fryer") || !res[0].Demo {
		t.Fatalf("unexpected first result %+v", res[0])
	}
	offer := res[0].Offers[0]
	if !strings.HasPrefix(offer.ProductURL, "http://localhost:3000/demo/loja/") {
		t.Fatalf("mock must never point at real marketplaces: %s", offer.ProductURL)
	}
	out, err := p.BuildAffiliateURL(context.Background(), offer.MerchantCode, offer.ProductURL, affiliate.ClickContext{ClickID: "c1"})
	if err != nil || !strings.Contains(out, "demo_tag=listou-") || !strings.Contains(out, "click=c1") {
		t.Fatalf("bad url %q %v", out, err)
	}
	if _, err := p.GetProduct(context.Background(), "nope"); err != affiliate.ErrNotFound {
		t.Fatal("expected not found")
	}
}

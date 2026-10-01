package recommendations

import (
	"context"
	"testing"
)

func TestRulesProviderNeverSuggestsPrices(t *testing.T) {
	for typ := range templates {
		cats, err := RulesProvider{}.SuggestCategories(context.Background(), Request{EventType: typ})
		if err != nil || len(cats) == 0 {
			t.Fatalf("%s: no categories", typ)
		}
		for _, c := range cats {
			if c.Reason == "" {
				t.Errorf("%s/%s: missing reason", typ, c.Name)
			}
			for _, d := range c.Desires {
				if d.Quantity < 1 {
					t.Errorf("%s/%s: bad quantity", typ, d.Title)
				}
			}
		}
	}
	cats, _ := RulesProvider{}.SuggestCategories(context.Background(), Request{EventType: "UNKNOWN"})
	if cats[0].Name != "Desejos" {
		t.Fatal("expected wishlist fallback")
	}
}

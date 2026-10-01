package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/listou/listou/apps/api/internal/affiliate"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
)

// linkAdapter is the registry code of the paste-a-link provider.
const linkAdapter = "LINK"

var (
	ErrInvalidLink    = httpx.BadRequest("INVALID_LINK", "Esse link não parece válido. Cole o endereço completo do produto, começando com https://.")
	ErrLinkUnreadable = httpx.BadRequest("LINK_UNREADABLE", "Não conseguimos ler os dados desse link. Digite o nome do item para continuar.")
)

// LinkPreview is what we could read from a pasted product link. Prices are
// intentionally absent: only the page's public title and image are used.
type LinkPreview struct {
	URL       string  `json:"url"`
	Title     string  `json:"title"`
	ImageURL  *string `json:"imageUrl"`
	StoreName string  `json:"storeName"`
	// Readable is false when the page gave no title (bot protection, no
	// metadata); the client then asks the user for the item name.
	Readable bool `json:"readable"`
}

func (s *Service) linkProduct(ctx context.Context, rawURL string) (affiliate.ProductData, error) {
	provider, ok := s.registry.Get(linkAdapter)
	if !ok {
		return affiliate.ProductData{}, httpx.NotFound("FEATURE_UNAVAILABLE", "Adicionar por link não está disponível.")
	}
	pd, err := provider.GetProduct(ctx, rawURL)
	if err != nil {
		if errors.Is(err, affiliate.ErrInvalidURL) {
			return affiliate.ProductData{}, ErrInvalidLink
		}
		return affiliate.ProductData{}, fmt.Errorf("link: %w", err)
	}
	return pd, nil
}

func (s *Service) PreviewLink(ctx context.Context, rawURL string) (LinkPreview, error) {
	pd, err := s.linkProduct(ctx, rawURL)
	if err != nil {
		return LinkPreview{}, err
	}
	p := LinkPreview{URL: pd.ExternalID, Title: pd.Title, ImageURL: pd.ImageURL, Readable: pd.Title != ""}
	if len(pd.Offers) > 0 && pd.Offers[0].StoreName != nil {
		p.StoreName = *pd.Offers[0].StoreName
	}
	return p, nil
}

// ImportLink stores a pasted link as a product with one offer. When the page
// is unreadable, the title typed by the user is required.
func (s *Service) ImportLink(ctx context.Context, rawURL, userTitle string) (Product, []Offer, error) {
	pd, err := s.linkProduct(ctx, rawURL)
	if err != nil {
		return Product{}, nil, err
	}
	userTitle = strings.Join(strings.Fields(userTitle), " ")
	if utf8.RuneCountInString(userTitle) > 200 {
		return Product{}, nil, httpx.Validation(map[string]string{"title": "Use até 200 caracteres."})
	}
	if pd.Title == "" {
		if userTitle == "" {
			return Product{}, nil, ErrLinkUnreadable
		}
		pd.Title = userTitle
		for i := range pd.Offers {
			pd.Offers[i].Title = userTitle
		}
	}
	return s.store(ctx, linkAdapter, pd)
}

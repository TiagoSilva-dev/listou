package curated

import (
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/platform/validate"
)

// Product is one curated catalog entry: a real product plus the affiliate
// links we generated for it in each partner's own panel.
type Product struct {
	ID         uuid.UUID `json:"id"`
	ExternalID string    `json:"externalId"`
	Title      string    `json:"title"`
	Brand      string    `json:"brand"`
	Category   string    `json:"category"`
	Keywords   string    `json:"keywords"`
	ImageURL   string    `json:"imageUrl"`
	Active     bool      `json:"active"`
	Offers     []Offer   `json:"offers"`
	// Emoji is derived from the category (legacy key rendered as a glyph).
	Emoji string `json:"-"`
}

type Offer struct {
	Merchant string `json:"merchant"`
	URL      string `json:"url"`
}

// Merchant is a store whose enabled provider is CURATED.
type Merchant struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Input is what the admin panel sends to create or replace a product.
type Input struct {
	Title    string  `json:"title"`
	Brand    string  `json:"brand"`
	Category string  `json:"category"`
	Keywords string  `json:"keywords"`
	ImageURL string  `json:"imageUrl"`
	Active   bool    `json:"active"`
	Offers   []Offer `json:"offers"`
}

var emojiByCategory = map[string]string{
	"cozinha": "🍳", "mesa posta": "🍽", "quarto": "🛏", "banheiro": "🛁",
	"sala": "🛋", "limpeza": "🧹", "organização": "📦", "organizacao": "📦",
	"bebê": "👶", "bebe": "👶", "viagem": "🧳", "lua de mel": "🧳",
}

// EmojiFor picks the legacy emoji key (rendered as a glyph) for a category.
func EmojiFor(category string) string {
	if e, ok := emojiByCategory[strings.ToLower(strings.TrimSpace(category))]; ok {
		return e
	}
	return "🎁"
}

func isHTTP(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// Validate trims the input and checks every field; offer merchants are
// checked against the allowed set by the caller.
func (in *Input) Validate(allowed map[string]bool) error {
	in.Title = strings.TrimSpace(in.Title)
	in.Brand = strings.TrimSpace(in.Brand)
	in.Category = strings.TrimSpace(in.Category)
	in.Keywords = strings.TrimSpace(in.Keywords)
	in.ImageURL = strings.TrimSpace(in.ImageURL)
	errs := validate.Errors{}
	errs.Length("title", in.Title, 3, 200, "Dê um nome ao produto (3 a 200 caracteres).")
	errs.Length("brand", in.Brand, 0, 80, "Marca com até 80 caracteres.")
	errs.Length("category", in.Category, 0, 60, "Categoria com até 60 caracteres.")
	errs.Length("keywords", in.Keywords, 0, 500, "Palavras-chave com até 500 caracteres.")
	if in.ImageURL != "" && (len(in.ImageURL) > 2000 || !isHTTP(in.ImageURL)) {
		errs.Add("imageUrl", "Use um endereço de imagem http(s) válido.")
	}
	if len(in.Offers) == 0 {
		errs.Add("offers", "Informe ao menos um link de afiliado.")
	}
	seen := map[string]bool{}
	for i, o := range in.Offers {
		in.Offers[i].URL = strings.TrimSpace(o.URL)
		o = in.Offers[i]
		switch {
		case !allowed[o.Merchant]:
			errs.Add("offers", "Loja inválida ou sem catálogo curado: "+o.Merchant+".")
		case seen[o.Merchant]:
			errs.Add("offers", "Só pode haver um link por loja.")
		case len(o.URL) > 2000 || !isHTTP(o.URL):
			errs.Add("offers", "O link de afiliado precisa ser um endereço http(s) válido.")
		}
		seen[o.Merchant] = true
	}
	return errs.Err()
}

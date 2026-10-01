// Package catalog owns products, offers and unified product discovery across
// affiliate providers.
package catalog

import (
	"time"

	"github.com/google/uuid"
)

type Merchant struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Product struct {
	ID             uuid.UUID `json:"id"`
	CanonicalTitle string    `json:"canonicalTitle"`
	Brand          *string   `json:"brand"`
	Description    *string   `json:"description"`
	ImageURL       *string   `json:"imageUrl"`
	Category       *string   `json:"category"`
	Emoji          *string   `json:"emoji"`
	Demo           bool      `json:"demo"`
}

// Offer is the public view of a ProductOffer. The merchant's product URL is
// deliberately absent: clients only ever get the /go/{id} redirect.
type Offer struct {
	ID                 uuid.UUID  `json:"id"`
	ProductID          uuid.UUID  `json:"-"`
	Merchant           Merchant   `json:"merchant"`
	Title              string     `json:"title"`
	PriceCents         *int64     `json:"priceCents"`
	OriginalPriceCents *int64     `json:"originalPriceCents"`
	Currency           string     `json:"currency"`
	Availability       string     `json:"availability"`
	ImageURL           *string    `json:"imageUrl"`
	GoURL              string     `json:"goUrl"`
	LastSyncedAt       *time.Time `json:"lastSyncedAt"`
	Demo               bool       `json:"demo"`
}

type SearchOffer struct {
	Merchant     Merchant `json:"merchant"`
	PriceCents   *int64   `json:"priceCents"`
	Availability string   `json:"availability"`
}

type SearchResult struct {
	ProviderCode     string        `json:"providerCode"`
	ExternalID       string        `json:"externalId"`
	Title            string        `json:"title"`
	Brand            *string       `json:"brand"`
	ImageURL         *string       `json:"imageUrl"`
	Emoji            *string       `json:"emoji"`
	Category         *string       `json:"category"`
	Rating           *float64      `json:"rating"`
	Demo             bool          `json:"demo"`
	Offers           []SearchOffer `json:"offers"`
	LowestPriceCents *int64        `json:"lowestPriceCents"`
	Match            *Ranked       `json:"match,omitempty"`
}

// Package lists manages gift lists, their categories and list items
// (desires). Items never depend on a marketplace: a product is optional.
package lists

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/catalog"
	"github.com/listou/listou/apps/api/internal/platform/validate"
)

type ItemStatus string

const (
	Available         ItemStatus = "AVAILABLE"
	PartiallyReserved ItemStatus = "PARTIALLY_RESERVED"
	Reserved          ItemStatus = "RESERVED"
	Purchased         ItemStatus = "PURCHASED"
	Archived          ItemStatus = "ARCHIVED"
)

// DeriveStatus computes availability from quantities (ADR-0007).
func DeriveStatus(desired, purchased, reserved int, archived bool) ItemStatus {
	switch {
	case archived:
		return Archived
	case purchased >= desired:
		return Purchased
	case purchased+reserved >= desired:
		return Reserved
	case purchased+reserved > 0:
		return PartiallyReserved
	default:
		return Available
	}
}

type GiftList struct {
	ID                      uuid.UUID `json:"id"`
	EventID                 uuid.UUID `json:"eventId"`
	Title                   string    `json:"title"`
	Description             *string   `json:"description"`
	AllowReservations       bool      `json:"allowReservations"`
	AllowGroupContributions bool      `json:"allowGroupContributions"`
}

type Category struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Emoji    *string   `json:"emoji"`
	Position int       `json:"position"`
}

type ProductSummary struct {
	ID             uuid.UUID `json:"id"`
	CanonicalTitle string    `json:"canonicalTitle"`
	Brand          *string   `json:"brand"`
	ImageURL       *string   `json:"imageUrl"`
	Emoji          *string   `json:"emoji"`
	Demo           bool      `json:"demo"`
}

type Item struct {
	ID                  uuid.UUID       `json:"id"`
	ListID              uuid.UUID       `json:"listId"`
	CategoryID          *uuid.UUID      `json:"categoryId"`
	ProductID           *uuid.UUID      `json:"-"`
	Title               string          `json:"title"`
	Description         *string         `json:"description"`
	Notes               *string         `json:"notes"`
	ImageURL            *string         `json:"imageUrl"`
	Emoji               *string         `json:"emoji"`
	ExternalURL         *string         `json:"externalUrl"`
	PriceReferenceCents *int64          `json:"priceReferenceCents"`
	Currency            string          `json:"currency"`
	Priority            string          `json:"priority"`
	DesiredQuantity     int             `json:"desiredQuantity"`
	PurchasedQuantity   int             `json:"purchasedQuantity"`
	ReservedQuantity    int             `json:"reservedQuantity"`
	AvailableQuantity   int             `json:"availableQuantity"`
	Status              ItemStatus      `json:"status"`
	Position            int             `json:"position"`
	Product             *ProductSummary `json:"product"`
	Offers              []catalog.Offer `json:"offers"`
	ArchivedAt          *time.Time      `json:"-"`
	CreatedAt           time.Time       `json:"createdAt"`
}

func (it *Item) finalize() {
	it.AvailableQuantity = max(0, it.DesiredQuantity-it.PurchasedQuantity-it.ReservedQuantity)
	it.Status = DeriveStatus(it.DesiredQuantity, it.PurchasedQuantity, it.ReservedQuantity, it.ArchivedAt != nil)
	if it.Offers == nil {
		it.Offers = []catalog.Offer{}
	}
}

var priorities = map[string]bool{"HIGH": true, "MEDIUM": true, "LOW": true}

type ItemInput struct {
	Title               *string `json:"title"`
	Description         *string `json:"description"`
	Notes               *string `json:"notes"`
	ImageURL            *string `json:"imageUrl"`
	Emoji               *string `json:"emoji"`
	ExternalURL         *string `json:"externalUrl"`
	PriceReferenceCents *int64  `json:"priceReferenceCents"`
	Priority            *string `json:"priority"`
	DesiredQuantity     *int    `json:"desiredQuantity"`
	PurchasedQuantity   *int    `json:"purchasedQuantity"`
	CategoryID          *string `json:"categoryId"`
	ProductID           *string `json:"productId"`
	Position            *int    `json:"position"`
}

// Validate checks fields present in the input. creating=true requires a title.
func (in *ItemInput) Validate(creating bool) error {
	errs := validate.Errors{}
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		in.Title = &t
	}
	if creating && (in.Title == nil || *in.Title == "") && in.ProductID == nil {
		errs.Add("title", "Dê um nome ao item.")
	}
	if in.Title != nil && (*in.Title != "" || !creating) {
		errs.Length("title", *in.Title, 1, 160, "Use de 1 a 160 caracteres.")
	}
	for field, v := range map[string]*string{"description": in.Description, "notes": in.Notes} {
		if v != nil {
			errs.Length(field, *v, 0, 1000, "Texto muito longo.")
		}
	}
	if in.Emoji != nil {
		errs.Length("emoji", *in.Emoji, 0, 8, "Emoji inválido.")
	}
	errs.OptionalURL("imageUrl", in.ImageURL)
	errs.OptionalURL("externalUrl", in.ExternalURL)
	if in.PriceReferenceCents != nil && (*in.PriceReferenceCents < 0 || *in.PriceReferenceCents > 100_000_000_00) {
		errs.Add("priceReferenceCents", "Valor inválido.")
	}
	if in.Priority != nil && !priorities[*in.Priority] {
		errs.Add("priority", "Prioridade inválida.")
	}
	if in.DesiredQuantity != nil && (*in.DesiredQuantity < 1 || *in.DesiredQuantity > 999) {
		errs.Add("desiredQuantity", "Quantidade entre 1 e 999.")
	}
	if in.PurchasedQuantity != nil && *in.PurchasedQuantity < 0 {
		errs.Add("purchasedQuantity", "Quantidade inválida.")
	}
	return errs.Err()
}

type CategoryInput struct {
	Name     *string `json:"name"`
	Emoji    *string `json:"emoji"`
	Position *int    `json:"position"`
}

func (in *CategoryInput) Validate(creating bool) error {
	errs := validate.Errors{}
	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		in.Name = &n
		errs.Length("name", n, 1, 60, "Use de 1 a 60 caracteres.")
	} else if creating {
		errs.Add("name", "Dê um nome à categoria.")
	}
	if in.Emoji != nil {
		errs.Length("emoji", *in.Emoji, 0, 8, "Emoji inválido.")
	}
	return errs.Err()
}

type ListInput struct {
	Title             *string `json:"title"`
	Description       *string `json:"description"`
	AllowReservations *bool   `json:"allowReservations"`
}

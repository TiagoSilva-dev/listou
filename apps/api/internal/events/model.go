// Package events manages occasions (Event), their settings and publication.
package events

import (
	"encoding/json"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/listou/listou/apps/api/internal/platform/validate"
)

var EventTypes = map[string]bool{
	"BABY_SHOWER": true, "WEDDING": true, "HOUSEWARMING": true, "BIRTHDAY": true,
	"GRADUATION": true, "TRAVEL": true, "CHRISTMAS": true, "WISHLIST": true, "CUSTOM": true,
}

var Themes = map[string]bool{"blush": true, "lavender": true, "sage": true, "sun": true, "night": true}

var defaultTheme = map[string]string{
	"BABY_SHOWER": "sage", "WEDDING": "lavender", "HOUSEWARMING": "blush", "BIRTHDAY": "sun",
	"GRADUATION": "night", "TRAVEL": "sun", "CHRISTMAS": "sage", "WISHLIST": "lavender", "CUSTOM": "blush",
}

var visibilities = map[string]bool{"PUBLIC": true, "UNLISTED": true, "PRIVATE": true}
var statuses = map[string]bool{"DRAFT": true, "PUBLISHED": true, "ARCHIVED": true}

type Event struct {
	ID                uuid.UUID
	OwnerID           uuid.UUID
	Type              string
	Title             string
	Slug              string
	Description       *string
	HostNames         *string
	EventDate         *time.Time
	Location          *string
	CoverImageURL     *string
	AvatarURL         *string
	Theme             string
	Visibility        string
	Status            string
	SurpriseMode      bool
	ShowReserverNames bool
	PublishedAt       *time.Time
	ListID            uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (e Event) MarshalJSON() ([]byte, error) {
	var date *string
	if e.EventDate != nil {
		s := e.EventDate.Format(time.DateOnly)
		date = &s
	}
	return json.Marshal(map[string]any{
		"id": e.ID, "type": e.Type, "title": e.Title, "slug": e.Slug,
		"description": e.Description, "hostNames": e.HostNames, "eventDate": date,
		"location": e.Location, "coverImageUrl": e.CoverImageURL, "avatarUrl": e.AvatarURL,
		"theme": e.Theme, "visibility": e.Visibility, "status": e.Status,
		"surpriseMode": e.SurpriseMode, "showReserverNames": e.ShowReserverNames,
		"publishedAt": e.PublishedAt, "listId": e.ListID,
		"createdAt": e.CreatedAt, "updatedAt": e.UpdatedAt,
	})
}

type CreateInput struct {
	Type          string  `json:"type"`
	Title         string  `json:"title"`
	HostNames     *string `json:"hostNames"`
	Description   *string `json:"description"`
	EventDate     *string `json:"eventDate"`
	Location      *string `json:"location"`
	CoverImageURL *string `json:"coverImageUrl"`
	Slug          *string `json:"slug"`
	Visibility    *string `json:"visibility"`
	Theme         *string `json:"theme"`
	// "SUGGESTED" seeds categories and desires from recommendations; "EMPTY" (default) does not.
	Template string `json:"template"`
}

type UpdateInput struct {
	Title             *string `json:"title"`
	HostNames         *string `json:"hostNames"`
	Description       *string `json:"description"`
	EventDate         *string `json:"eventDate"`
	Location          *string `json:"location"`
	CoverImageURL     *string `json:"coverImageUrl"`
	Slug              *string `json:"slug"`
	Visibility        *string `json:"visibility"`
	Status            *string `json:"status"`
	Theme             *string `json:"theme"`
	SurpriseMode      *bool   `json:"surpriseMode"`
	ShowReserverNames *bool   `json:"showReserverNames"`
}

func parseDate(errs validate.Errors, s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.DateOnly, *s)
	if err != nil {
		errs.Add("eventDate", "Data inválida.")
		return nil
	}
	return &t
}

func validateOptional(errs validate.Errors, field string, s *string, max int) {
	if s != nil {
		errs.Length(field, *s, 0, max, "Texto muito longo.")
	}
}

// Validate normalizes and validates a creation request, returning the parsed date.
func (in *CreateInput) Validate() (*time.Time, error) {
	errs := validate.Errors{}
	in.Title = strings.TrimSpace(in.Title)
	in.HostNames, in.Description, in.Location = validate.Trim(in.HostNames), validate.Trim(in.Description), validate.Trim(in.Location)
	in.CoverImageURL, in.Slug = validate.Trim(in.CoverImageURL), validate.Trim(in.Slug)
	if !EventTypes[in.Type] {
		errs.Add("type", "Escolha o tipo do momento.")
	}
	errs.Length("title", in.Title, 1, 120, "Dê um nome ao momento (até 120 caracteres).")
	validateOptional(errs, "hostNames", in.HostNames, 120)
	validateOptional(errs, "description", in.Description, 2000)
	validateOptional(errs, "location", in.Location, 160)
	errs.OptionalURL("coverImageUrl", in.CoverImageURL)
	if in.Slug != nil && !ValidSlug(*in.Slug) {
		errs.Add("slug", "Use de 3 a 60 letras minúsculas, números e hífens.")
	}
	if in.Visibility != nil && !visibilities[*in.Visibility] {
		errs.Add("visibility", "Visibilidade inválida.")
	}
	if in.Theme != nil && !Themes[*in.Theme] {
		errs.Add("theme", "Tema inválido.")
	}
	if in.Template != "" && in.Template != "SUGGESTED" && in.Template != "EMPTY" {
		errs.Add("template", "Modelo inválido.")
	}
	date := parseDate(errs, in.EventDate)
	return date, errs.Err()
}

func (in *UpdateInput) Validate() (*time.Time, error) {
	errs := validate.Errors{}
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		in.Title = &t
		errs.Length("title", t, 1, 120, "Dê um nome ao momento (até 120 caracteres).")
	}
	validateOptional(errs, "hostNames", in.HostNames, 120)
	validateOptional(errs, "description", in.Description, 2000)
	validateOptional(errs, "location", in.Location, 160)
	errs.OptionalURL("coverImageUrl", in.CoverImageURL)
	if in.Slug != nil && !ValidSlug(*in.Slug) {
		errs.Add("slug", "Use de 3 a 60 letras minúsculas, números e hífens.")
	}
	if in.Visibility != nil && !visibilities[*in.Visibility] {
		errs.Add("visibility", "Visibilidade inválida.")
	}
	if in.Status != nil && !statuses[*in.Status] {
		errs.Add("status", "Status inválido.")
	}
	if in.Theme != nil && !Themes[*in.Theme] {
		errs.Add("theme", "Tema inválido.")
	}
	date := parseDate(errs, in.EventDate)
	return date, errs.Err()
}

var reservedSlugs = map[string]bool{
	"admin": true, "api": true, "app": true, "dashboard": true, "entrar": true, "criar": true,
	"criar-conta": true, "go": true, "listou": true, "suporte": true, "ajuda": true, "demo": true,
}

func ValidSlug(s string) bool {
	if len(s) < 3 || len(s) > 60 || reservedSlugs[s] {
		return false
	}
	prevDash := true
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			prevDash = false
		case r == '-' && !prevDash:
			prevDash = true
		default:
			return false
		}
	}
	return !prevDash
}

// Slugify turns "Tiago & Júlia" into "tiago-e-julia".
func Slugify(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	plain, _, err := transform.String(t, strings.ToLower(s))
	if err != nil {
		plain = strings.ToLower(s)
	}
	plain = strings.ReplaceAll(plain, "&", " e ")
	var b strings.Builder
	dash := true
	for _, r := range plain {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash {
			b.WriteByte('-')
			dash = true
		}
		if b.Len() >= 60 {
			break
		}
	}
	return strings.Trim(b.String(), "-")
}

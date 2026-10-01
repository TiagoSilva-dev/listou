// Package publiclists serves the read-only public registry page.
package publiclists

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/lists"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
)

var ErrNotFound = httpx.NotFound("LIST_NOT_FOUND", "Esta lista não foi encontrada.")

type Service struct {
	db    database.DBTX
	lists *lists.Service
}

func NewService(db database.DBTX, l *lists.Service) *Service { return &Service{db: db, lists: l} }

type publicEvent struct {
	Type          string  `json:"type"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	Description   *string `json:"description"`
	HostNames     *string `json:"hostNames"`
	EventDate     *string `json:"eventDate"`
	Location      *string `json:"location"`
	CoverImageURL *string `json:"coverImageUrl"`
	AvatarURL     *string `json:"avatarUrl"`
	Theme         string  `json:"theme"`
	Visibility    string  `json:"visibility"`
	Status        string  `json:"status"`
}

type publicItem struct {
	ID                  uuid.UUID             `json:"id"`
	CategoryID          *uuid.UUID            `json:"categoryId"`
	Title               string                `json:"title"`
	Description         *string               `json:"description"`
	ImageURL            *string               `json:"imageUrl"`
	Emoji               *string               `json:"emoji"`
	ExternalURL         *string               `json:"externalUrl"`
	PriceReferenceCents *int64                `json:"priceReferenceCents"`
	Currency            string                `json:"currency"`
	Priority            string                `json:"priority"`
	DesiredQuantity     int                   `json:"desiredQuantity"`
	PurchasedQuantity   int                   `json:"purchasedQuantity"`
	ReservedQuantity    int                   `json:"reservedQuantity"`
	AvailableQuantity   int                   `json:"availableQuantity"`
	Status              lists.ItemStatus      `json:"status"`
	Product             *lists.ProductSummary `json:"product"`
	Offers              any                   `json:"offers"`
}

type Progress struct {
	TotalUnits     int `json:"totalUnits"`
	ReservedUnits  int `json:"reservedUnits"`
	PurchasedUnits int `json:"purchasedUnits"`
	AvailableUnits int `json:"availableUnits"`
}

type View struct {
	Event      publicEvent      `json:"event"`
	List       map[string]any   `json:"list"`
	Categories []lists.Category `json:"categories"`
	Items      []publicItem     `json:"items"`
	Progress   Progress         `json:"progress"`
	Preview    bool             `json:"preview"`
}

// Get returns the public view. Drafts are visible only to managers (preview).
// PRIVATE lists are never served publicly.
func (s *Service) Get(ctx context.Context, slug string, viewer *auth.User) (View, error) {
	var e struct {
		id        uuid.UUID
		ownerID   uuid.UUID
		ev        publicEvent
		date      *time.Time
		published bool
	}
	err := s.db.QueryRow(ctx, `
		SELECT id, owner_id, type, title, slug, description, host_names, event_date, location, cover_image_url,
			avatar_url, theme, visibility, status
		FROM events WHERE slug = $1 AND deleted_at IS NULL`, slug).Scan(&e.id, &e.ownerID, &e.ev.Type, &e.ev.Title,
		&e.ev.Slug, &e.ev.Description, &e.ev.HostNames, &e.date, &e.ev.Location, &e.ev.CoverImageURL,
		&e.ev.AvatarURL, &e.ev.Theme, &e.ev.Visibility, &e.ev.Status)
	if err != nil {
		if database.IsNoRows(err) {
			return View{}, ErrNotFound
		}
		return View{}, err
	}
	manager := false
	if viewer != nil {
		if viewer.ID == e.ownerID {
			manager = true
		} else {
			_ = s.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM event_members WHERE event_id = $1 AND user_id = $2 AND role = 'CO_OWNER')`,
				e.id, viewer.ID).Scan(&manager)
		}
	}
	live := e.ev.Status == "PUBLISHED" && e.ev.Visibility != "PRIVATE"
	if !live && !manager {
		return View{}, ErrNotFound
	}
	if e.date != nil {
		d := e.date.Format(time.DateOnly)
		e.ev.EventDate = &d
	}
	l, err := s.lists.PrimaryList(ctx, s.db, e.id)
	if err != nil {
		return View{}, err
	}
	lv, err := s.lists.View(ctx, s.db, l)
	if err != nil {
		return View{}, err
	}
	out := View{
		Event:      e.ev,
		List:       map[string]any{"id": l.ID, "title": l.Title, "allowReservations": l.AllowReservations},
		Categories: lv.Categories,
		Items:      make([]publicItem, 0, len(lv.Items)),
		Preview:    !live,
	}
	for _, it := range lv.Items {
		out.Items = append(out.Items, publicItem{
			ID: it.ID, CategoryID: it.CategoryID, Title: it.Title, Description: it.Description, ImageURL: it.ImageURL,
			Emoji: it.Emoji, ExternalURL: it.ExternalURL, PriceReferenceCents: it.PriceReferenceCents,
			Currency: it.Currency, Priority: it.Priority, DesiredQuantity: it.DesiredQuantity,
			PurchasedQuantity: it.PurchasedQuantity, ReservedQuantity: it.ReservedQuantity,
			AvailableQuantity: it.AvailableQuantity, Status: it.Status, Product: it.Product, Offers: it.Offers,
		})
		out.Progress.TotalUnits += it.DesiredQuantity
		out.Progress.ReservedUnits += it.ReservedQuantity
		out.Progress.PurchasedUnits += it.PurchasedQuantity
		out.Progress.AvailableUnits += it.AvailableQuantity
	}
	return out, nil
}

type Handler struct {
	svc      *Service
	optional func(*http.Request) *auth.User
}

func NewHandler(svc *Service, optional func(*http.Request) *auth.User) *Handler {
	return &Handler{svc: svc, optional: optional}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/public/lists/{slug}", h.get)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Get(r.Context(), r.PathValue("slug"), h.optional(r))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	// Live pages are CDN-cacheable briefly; previews are per-user.
	if v.Preview || r.Header.Get("Cookie") != "" {
		w.Header().Set("Cache-Control", "private, no-store")
	}
	httpx.JSON(w, http.StatusOK, v)
}

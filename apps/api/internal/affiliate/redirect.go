package affiliate

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/listou/listou/apps/api/internal/analytics"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
	"github.com/listou/listou/apps/api/internal/platform/logger"
)

var ErrOfferNotFound = httpx.NotFound("OFFER_NOT_FOUND", "Oferta não encontrada.")

// Redirector implements GET /go/{offerId}: resolve the offer, build the
// allowed affiliate URL through its provider, record the click, 302.
type Redirector struct {
	pool     *pgxpool.Pool
	registry *Registry
	tracker  *analytics.Recorder
	secure   bool
}

func NewRedirector(pool *pgxpool.Pool, reg *Registry, tracker *analytics.Recorder, secureCookies bool) *Redirector {
	return &Redirector{pool: pool, registry: reg, tracker: tracker, secure: secureCookies}
}

func (rd *Redirector) Register(mux *http.ServeMux, limit httpx.Middleware) {
	mux.Handle("GET /go/{offerId}", limit(http.HandlerFunc(rd.handle)))
}

type offerRow struct {
	offerID      uuid.UUID
	merchantID   uuid.UUID
	merchantCode string
	productURL   string
	productID    uuid.UUID
	price        *int64
	adapter      string
}

func (rd *Redirector) handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	offerID, ok := ids.Parse(r.PathValue("offerId"))
	if !ok {
		httpx.Fail(w, r, ErrOfferNotFound)
		return
	}
	var o offerRow
	err := rd.pool.QueryRow(ctx, `
		SELECT o.id, m.id, m.code, o.product_url, o.product_id, o.price_cents, p.adapter
		FROM product_offers o
		JOIN merchants m ON m.id = o.merchant_id AND m.active
		JOIN affiliate_providers p ON p.merchant_id = m.id AND p.enabled
		WHERE o.id = $1`, offerID).Scan(&o.offerID, &o.merchantID, &o.merchantCode, &o.productURL, &o.productID, &o.price, &o.adapter)
	if err != nil {
		if database.IsNoRows(err) {
			httpx.Fail(w, r, ErrOfferNotFound)
			return
		}
		httpx.Fail(w, r, fmt.Errorf("go: lookup offer: %w", err))
		return
	}
	provider, ok := rd.registry.Get(o.adapter)
	if !ok {
		httpx.Fail(w, r, fmt.Errorf("go: no adapter %q", o.adapter))
		return
	}

	q := r.URL.Query()
	eventID, itemID := rd.resolveContext(ctx, o.productID, q.Get("item"))
	clickID := ids.New()
	target, err := provider.BuildAffiliateURL(ctx, o.merchantCode, o.productURL, ClickContext{
		ClickID: clickID.String(), UTMSource: q.Get("utm_source"), UTMMedium: q.Get("utm_medium"), UTMCampaign: q.Get("utm_campaign"),
	})
	if err != nil || !safeTarget(target) {
		httpx.Fail(w, r, fmt.Errorf("go: build url: %v", err))
		return
	}

	visitor := analytics.EnsureVisitor(w, r, rd.secure)
	if err := rd.recordClick(ctx, clickID, o, eventID, itemID, visitor, r); err != nil {
		// Never block the guest on analytics failure.
		logger.From(ctx).Warn("click not recorded", "error", err.Error())
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	http.Redirect(w, r, target, http.StatusFound)
}

// resolveContext attaches the click to an event/item only if the item really
// uses this offer's product (never trusting the query string blindly).
func (rd *Redirector) resolveContext(ctx context.Context, productID uuid.UUID, rawItem string) (*uuid.UUID, *uuid.UUID) {
	itemID, ok := ids.Parse(rawItem)
	if !ok {
		return nil, nil
	}
	var eventID uuid.UUID
	err := rd.pool.QueryRow(ctx, `
		SELECT g.event_id FROM list_items i JOIN gift_lists g ON g.id = i.gift_list_id
		WHERE i.id = $1 AND i.product_id = $2`, itemID, productID).Scan(&eventID)
	if err != nil {
		return nil, nil
	}
	return &eventID, &itemID
}

func (rd *Redirector) recordClick(ctx context.Context, clickID uuid.UUID, o offerRow, eventID, itemID *uuid.UUID, visitor uuid.UUID, r *http.Request) error {
	q := r.URL.Query()
	_, err := rd.pool.Exec(ctx, `
		INSERT INTO click_events (id, offer_id, merchant_id, event_id, list_item_id, visitor_id, referrer_host,
			utm_source, utm_medium, utm_campaign, device_class, price_cents)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		clickID, o.offerID, o.merchantID, eventID, itemID, visitor, referrerHost(r.Referer()),
		clip(q.Get("utm_source")), clip(q.Get("utm_medium")), clip(q.Get("utm_campaign")), deviceClass(r.UserAgent()), o.price)
	if err != nil {
		return err
	}
	rd.tracker.Track(ctx, analytics.Event{Name: analytics.OutboundClicked, EventID: eventID, ItemID: itemID, VisitorID: &visitor,
		Props: map[string]any{"merchant": o.merchantCode}})
	return nil
}

func safeTarget(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func clip(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if len(s) > 80 {
		s = s[:80]
	}
	return &s
}

// referrerHost keeps only the host: paths and queries can carry personal data.
func referrerHost(ref string) *string {
	if ref == "" {
		return nil
	}
	u, err := url.Parse(ref)
	if err != nil || u.Hostname() == "" {
		return nil
	}
	h := u.Hostname()
	return &h
}

// deviceClass is a coarse bucket derived from the UA; the UA itself is never stored.
func deviceClass(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet"):
		return "TABLET"
	case strings.Contains(ua, "mobi") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone"):
		return "MOBILE"
	case ua == "":
		return "UNKNOWN"
	default:
		return "DESKTOP"
	}
}

// Package analytics stores first-party product events and outbound clicks.
// It is privacy-first: random visitor ids, no IPs, no fingerprinting.
package analytics

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/ids"
	"github.com/listou/listou/apps/api/internal/platform/logger"
)

const (
	ListCreated     = "LIST_CREATED"
	ListViewed      = "LIST_VIEWED"
	ItemCreated     = "ITEM_CREATED"
	ItemViewed      = "ITEM_VIEWED"
	ProductSearched = "PRODUCT_SEARCHED"
	OfferViewed     = "OFFER_VIEWED"
	OutboundClicked = "OUTBOUND_CLICKED"
	ItemReserved    = "ITEM_RESERVED"
	ItemUnreserved  = "ITEM_UNRESERVED"
	ItemPurchased   = "ITEM_PURCHASED"
	ShareCreated    = "SHARE_CREATED"
	ShareOpened     = "SHARE_OPENED"
	UserRegistered  = "USER_REGISTERED"
	AISuggested     = "AI_SUGGESTIONS_REQUESTED"
	AIApplied       = "AI_SUGGESTIONS_APPLIED"
)

// clientEvents are the only names browsers may send to /analytics/track;
// everything else is recorded server-side where it actually happens.
var clientEvents = map[string]bool{
	ListViewed:   true,
	ItemViewed:   true,
	OfferViewed:  true,
	ShareCreated: true,
	ShareOpened:  true,
}

type Event struct {
	Name      string
	EventID   *uuid.UUID
	ItemID    *uuid.UUID
	UserID    *uuid.UUID
	VisitorID *uuid.UUID
	Props     map[string]any
}

type Recorder struct {
	db database.DBTX
}

func NewRecorder(db database.DBTX) *Recorder { return &Recorder{db: db} }

// Track stores an event. Analytics must never break the user flow, so
// failures are logged rather than returned.
func (r *Recorder) Track(ctx context.Context, e Event) {
	if e.Props == nil {
		e.Props = map[string]any{}
	}
	raw, err := json.Marshal(e.Props)
	if err == nil {
		// Detach from request cancellation but keep a short deadline.
		c, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_, err = r.db.Exec(c, `
			INSERT INTO analytics_events (id, name, event_id, list_item_id, user_id, visitor_id, properties)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			ids.New(), e.Name, e.EventID, e.ItemID, e.UserID, e.VisitorID, raw)
	}
	if err != nil {
		logger.From(ctx).Warn("analytics track failed", "event", e.Name, "error", err.Error())
	}
}

// TrackUser satisfies auth.Tracker.
func (r *Recorder) TrackUser(ctx context.Context, name, userID string, props map[string]any) {
	id, ok := ids.Parse(userID)
	if !ok {
		return
	}
	r.Track(ctx, Event{Name: name, UserID: &id, Props: props})
}

const VisitorCookie = "listou_vid"

// VisitorID reads the first-party random visitor id, if present and valid.
func VisitorID(r *http.Request) *uuid.UUID {
	c, err := r.Cookie(VisitorCookie)
	if err != nil {
		return nil
	}
	id, ok := ids.Parse(c.Value)
	if !ok {
		return nil
	}
	return &id
}

// EnsureVisitor returns the visitor id, minting and setting one if absent.
func EnsureVisitor(w http.ResponseWriter, r *http.Request, secure bool) uuid.UUID {
	if id := VisitorID(r); id != nil {
		return *id
	}
	id := uuid.New()
	http.SetCookie(w, &http.Cookie{
		Name:     VisitorCookie,
		Value:    id.String(),
		Path:     "/",
		MaxAge:   365 * 24 * 3600,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	return id
}

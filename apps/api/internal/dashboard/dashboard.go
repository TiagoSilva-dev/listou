// Package dashboard aggregates creator-facing metrics for an event.
package dashboard

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/access"
	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/events"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

type Stats struct {
	DaysRemaining  *int `json:"daysRemaining"`
	ItemsCount     int  `json:"itemsCount"`
	TotalUnits     int  `json:"totalUnits"`
	ReservedUnits  int  `json:"reservedUnits"`
	PurchasedUnits int  `json:"purchasedUnits"`
	Views          int  `json:"views"`
	OutboundClicks int  `json:"outboundClicks"`
}

type Activity struct {
	Kind      string    `json:"kind"`
	ItemTitle *string   `json:"itemTitle"`
	GuestName *string   `json:"guestName"`
	At        time.Time `json:"at"`
}

type View struct {
	Event          events.Event `json:"event"`
	Stats          Stats        `json:"stats"`
	SurpriseMode   bool         `json:"surpriseMode"`
	RecentActivity []Activity   `json:"recentActivity"`
}

type Service struct {
	db     database.DBTX
	events *events.Service
	now    func() time.Time
}

func NewService(db database.DBTX, ev *events.Service) *Service {
	return &Service{db: db, events: ev, now: time.Now}
}

// Get builds the dashboard. In surprise mode the creator sees only
// aggregate counts: no item titles and no guest names in the activity feed.
func (s *Service) Get(ctx context.Context, userID, eventID uuid.UUID) (View, error) {
	ev, err := s.events.Get(ctx, userID, eventID)
	if err != nil {
		return View{}, err
	}
	v := View{Event: ev, SurpriseMode: ev.SurpriseMode, RecentActivity: []Activity{}}
	if ev.EventDate != nil {
		today := time.Date(s.now().Year(), s.now().Month(), s.now().Day(), 0, 0, 0, 0, time.UTC)
		days := int(ev.EventDate.Sub(today).Hours() / 24)
		v.Stats.DaysRemaining = &days
	}
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(desired_quantity), 0), COALESCE(SUM(reserved_quantity), 0), COALESCE(SUM(purchased_quantity), 0)
		FROM list_items WHERE gift_list_id = $1 AND archived_at IS NULL`, ev.ListID).
		Scan(&v.Stats.ItemsCount, &v.Stats.TotalUnits, &v.Stats.ReservedUnits, &v.Stats.PurchasedUnits)
	if err != nil {
		return View{}, err
	}
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE name = 'LIST_VIEWED'), COUNT(*) FILTER (WHERE name = 'OUTBOUND_CLICKED')
		FROM analytics_events WHERE event_id = $1`, eventID).Scan(&v.Stats.Views, &v.Stats.OutboundClicks)
	if err != nil {
		return View{}, err
	}

	rows, err := s.db.Query(ctx, `
		(SELECT CASE WHEN r.status = 'CANCELLED' THEN 'CANCELLED' WHEN r.kind = 'PURCHASE' OR r.status = 'CONFIRMED' THEN 'PURCHASED' ELSE 'RESERVED' END,
			i.title, r.guest_name, r.created_at
		 FROM reservations r JOIN list_items i ON i.id = r.list_item_id JOIN gift_lists g ON g.id = i.gift_list_id
		 WHERE g.event_id = $1 ORDER BY r.created_at DESC LIMIT 10)
		UNION ALL
		(SELECT 'CLICKED', i.title, NULL, c.created_at
		 FROM click_events c LEFT JOIN list_items i ON i.id = c.list_item_id
		 WHERE c.event_id = $1 ORDER BY c.created_at DESC LIMIT 10)
		ORDER BY 4 DESC LIMIT 10`, eventID)
	if err != nil {
		return View{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var a Activity
		if err := rows.Scan(&a.Kind, &a.ItemTitle, &a.GuestName, &a.At); err != nil {
			return View{}, err
		}
		if ev.SurpriseMode {
			a.ItemTitle, a.GuestName = nil, nil
		}
		v.RecentActivity = append(v.RecentActivity, a)
	}
	return v, rows.Err()
}

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(mux *http.ServeMux, require func(http.HandlerFunc) http.Handler) {
	mux.Handle("GET /api/v1/events/{id}/dashboard", require(h.get))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		httpx.Fail(w, r, access.ErrEventNotFound)
		return
	}
	v, err := h.svc.Get(r.Context(), u.ID, id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, v)
}

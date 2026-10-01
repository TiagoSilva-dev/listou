package analytics

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

type Handler struct {
	rec    *Recorder
	db     database.DBTX
	secure bool
}

func NewHandler(rec *Recorder, db database.DBTX, secureCookies bool) *Handler {
	return &Handler{rec: rec, db: db, secure: secureCookies}
}

func (h *Handler) Register(mux *http.ServeMux, limit httpx.Middleware) {
	mux.Handle("POST /api/v1/analytics/track", limit(http.HandlerFunc(h.track)))
}

type trackInput struct {
	Name       string         `json:"name"`
	Slug       string         `json:"slug"`
	ItemID     string         `json:"itemId"`
	Properties map[string]any `json:"properties"`
}

func (h *Handler) track(w http.ResponseWriter, r *http.Request) {
	var in trackInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if !clientEvents[in.Name] {
		httpx.Fail(w, r, httpx.BadRequest("UNKNOWN_EVENT", "Evento de analytics desconhecido."))
		return
	}
	props, ok := sanitizeProps(in.Properties)
	if !ok {
		httpx.Fail(w, r, httpx.BadRequest("INVALID_PROPERTIES", "Propriedades inválidas."))
		return
	}
	e := Event{Name: in.Name, Props: props}
	if in.Slug != "" {
		var eventID uuid.UUID
		err := h.db.QueryRow(r.Context(), `
			SELECT id FROM events
			WHERE slug = $1 AND status = 'PUBLISHED' AND visibility <> 'PRIVATE' AND deleted_at IS NULL`, in.Slug).Scan(&eventID)
		if err != nil {
			if database.IsNoRows(err) {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			httpx.Fail(w, r, err)
			return
		}
		e.EventID = &eventID
		if id, ok := ids.Parse(in.ItemID); ok {
			// Only attach items that belong to this event.
			var exists bool
			_ = h.db.QueryRow(r.Context(), `
				SELECT EXISTS (SELECT 1 FROM list_items i JOIN gift_lists l ON l.id = i.gift_list_id
				WHERE i.id = $1 AND l.event_id = $2)`, id, eventID).Scan(&exists)
			if exists {
				e.ItemID = &id
			}
		}
	}
	vid := EnsureVisitor(w, r, h.secure)
	e.VisitorID = &vid
	h.rec.Track(r.Context(), e)
	w.WriteHeader(http.StatusAccepted)
}

// sanitizeProps keeps a small flat map of scalar values.
func sanitizeProps(in map[string]any) (map[string]any, bool) {
	if len(in) > 10 {
		return nil, false
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		if len(k) > 40 {
			return nil, false
		}
		switch val := v.(type) {
		case string:
			if len(val) > 200 {
				return nil, false
			}
			out[k] = val
		case float64, bool:
			out[k] = val
		default:
			return nil, false
		}
	}
	return out, true
}

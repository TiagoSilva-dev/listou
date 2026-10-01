package reservations

import (
	"net/http"

	"github.com/listou/listou/apps/api/internal/analytics"
	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

const manageTokenHeader = "X-Reservation-Token"

type Handler struct {
	svc    *Service
	secure bool
}

func NewHandler(svc *Service, secureCookies bool) *Handler {
	return &Handler{svc: svc, secure: secureCookies}
}

func (h *Handler) Register(mux *http.ServeMux, limit httpx.Middleware, require func(http.HandlerFunc) http.Handler) {
	mux.Handle("POST /api/v1/public/lists/{slug}/items/{itemId}/reservations", limit(http.HandlerFunc(h.create)))
	mux.Handle("DELETE /api/v1/reservations/{id}", limit(http.HandlerFunc(h.cancel)))
	mux.Handle("POST /api/v1/reservations/{id}/confirm", limit(http.HandlerFunc(h.confirm)))
	mux.Handle("DELETE /api/v1/manage/reservations/{id}", require(h.cancelByOwner))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	itemID, ok := ids.Parse(r.PathValue("itemId"))
	if !ok {
		httpx.Fail(w, r, ErrItemNotFound)
		return
	}
	var in CreateInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	vid := analytics.EnsureVisitor(w, r, h.secure)
	res, err := h.svc.Create(r.Context(), r.PathValue("slug"), itemID, &vid, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"reservation": res})
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		httpx.Fail(w, r, ErrReservationMissing)
		return
	}
	if err := h.svc.Cancel(r.Context(), id, r.Header.Get(manageTokenHeader)); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) confirm(w http.ResponseWriter, r *http.Request) {
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		httpx.Fail(w, r, ErrReservationMissing)
		return
	}
	if err := h.svc.Confirm(r.Context(), id, r.Header.Get(manageTokenHeader)); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) cancelByOwner(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		httpx.Fail(w, r, ErrReservationMissing)
		return
	}
	if err := h.svc.CancelByOwner(r.Context(), u.ID, id); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

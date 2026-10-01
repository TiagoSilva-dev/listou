package admin

import (
	"net/http"

	"github.com/listou/listou/apps/api/internal/affiliate/curated"
	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register wires the admin routes; requireAdmin must reject non-ADMIN users.
func (h *Handler) Register(mux *http.ServeMux, requireAdmin func(http.HandlerFunc) http.Handler) {
	mux.Handle("GET /api/v1/admin/curated-products", requireAdmin(h.list))
	mux.Handle("POST /api/v1/admin/curated-products", requireAdmin(h.create))
	mux.Handle("PUT /api/v1/admin/curated-products/{id}", requireAdmin(h.update))
	mux.Handle("DELETE /api/v1/admin/curated-products/{id}", requireAdmin(h.delete))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	products, merchants, err := h.svc.List(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"products": products, "merchants": merchants})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in curated.Input
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	u, _ := auth.UserFrom(r.Context())
	p, err := h.svc.Create(r.Context(), u, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"product": p})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		httpx.Fail(w, r, errProductNotFound)
		return
	}
	var in curated.Input
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	u, _ := auth.UserFrom(r.Context())
	p, err := h.svc.Update(r.Context(), u, id, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"product": p})
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		httpx.Fail(w, r, errProductNotFound)
		return
	}
	u, _ := auth.UserFrom(r.Context())
	if err := h.svc.Delete(r.Context(), u, id); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

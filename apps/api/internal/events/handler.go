package events

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/access"
	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(mux *http.ServeMux, require func(http.HandlerFunc) http.Handler) {
	mux.Handle("POST /api/v1/events", require(h.create))
	mux.Handle("GET /api/v1/events", require(h.list))
	mux.Handle("GET /api/v1/events/{id}", require(h.get))
	mux.Handle("PATCH /api/v1/events/{id}", require(h.update))
	mux.Handle("DELETE /api/v1/events/{id}", require(h.delete))
	mux.Handle("GET /api/v1/slugs/{slug}", require(h.slug))
}

func eventID(r *http.Request) (uuid.UUID, error) {
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		return uuid.Nil, access.ErrEventNotFound
	}
	return id, nil
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	var in CreateInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	e, err := h.svc.Create(r.Context(), u.ID, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"event": e})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	out, err := h.svc.List(r.Context(), u.ID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"events": out})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	id, err := eventID(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	e, err := h.svc.Get(r.Context(), u.ID, id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"event": e})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	id, err := eventID(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in UpdateInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	e, err := h.svc.Update(r.Context(), u.ID, id, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"event": e})
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	id, err := eventID(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.Delete(r.Context(), u.ID, id); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) slug(w http.ResponseWriter, r *http.Request) {
	ok, err := h.svc.SlugAvailable(r.Context(), r.PathValue("slug"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"available": ok})
}

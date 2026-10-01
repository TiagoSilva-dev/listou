package catalog

import (
	"net/http"
	"strconv"

	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(mux *http.ServeMux, require func(http.HandlerFunc) http.Handler) {
	mux.Handle("GET /api/v1/products/search", require(h.search))
	mux.Handle("POST /api/v1/products/import", require(h.importProduct))
	mux.Handle("GET /api/v1/products/{id}", require(h.get))
	mux.Handle("GET /api/v1/products/{id}/offers", require(h.offers))
}

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 40 {
		limit = 24
	}
	res, err := h.svc.Search(r.Context(), SearchParams{
		Query: q.Get("q"), Sort: q.Get("sort"), Category: q.Get("category"), Limit: limit, UserID: &u.ID,
	})
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if res == nil {
		res = []SearchResult{}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"results": res})
}

type importInput struct {
	ProviderCode string `json:"providerCode"`
	ExternalID   string `json:"externalId"`
}

func (h *Handler) importProduct(w http.ResponseWriter, r *http.Request) {
	var in importInput
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	p, offers, err := h.svc.Import(r.Context(), in.ProviderCode, in.ExternalID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"product": p, "offers": offers})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		httpx.Fail(w, r, ErrProductNotFound)
		return
	}
	p, offers, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"product": p, "offers": offers})
}

func (h *Handler) offers(w http.ResponseWriter, r *http.Request) {
	id, ok := ids.Parse(r.PathValue("id"))
	if !ok {
		httpx.Fail(w, r, ErrProductNotFound)
		return
	}
	_, offers, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"offers": offers})
}
